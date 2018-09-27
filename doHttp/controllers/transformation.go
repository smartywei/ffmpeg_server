package controllers

import (
	"net/http"
	"fmt"
	"strconv"
	"time"
	"encoding/json"
	"ffmpeg/drive"
	"ffmpeg/doHttp/tools"
	"ffmpeg/doHttp/tools/config"
	"ffmpeg/doHttp/tools/fileUtil"
	"os"
	"encoding/base64"
	"strings"
	"net"
)

//开始处理转换请求。将请求加入redis队列
func StartTransformation(w http.ResponseWriter, r *http.Request) {

	var size string

	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		fmt.Fprintln(w, "request method is not support")
		return
	}

	r.ParseForm()                   //解析参数
	r.ParseMultipartForm(1024 * 10) //解析form-data 格式

	transformationIPs, err := config.Config("transformationIPs")

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(500)
		fmt.Fprintln(w, "transformationIPs is error")
		return
	}

	//这个路由需要限制ip

	var nowIp string

	nowIp = string([]byte(r.RemoteAddr)[0:strings.LastIndex(r.RemoteAddr, ":")])

	transformationIPsArr := strings.Split(transformationIPs.(string), ",")

	var ipCanAccess = false

	if nowIp != "127.0.0.1" {
		for _, v := range transformationIPsArr {
			if nowIp == v {
				ipCanAccess = true
				break
			}
		}
	} else {
		nowIp = r.Header.Get("X-Real-IP")
		nowIp = net.ParseIP(nowIp).String()

		for _, v := range transformationIPsArr {
			if nowIp == v {
				ipCanAccess = true
				break
			}
		}
	}

	if !ipCanAccess {
		w.WriteHeader(403)
		return
	}

	sourceHref := r.Form["href"]

	if len(sourceHref) <= 0 || len(sourceHref[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "Href can't is null")
		return
	}

	sourceType := r.Form["type"]

	if len(sourceType) <= 0 || len(sourceType[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "type can't is null")
		return
	}

	sourceSize := r.Form["size"]

	if len(sourceSize) <= 0 || len(sourceSize[0]) <= 0 {

		headHref, err := base64.StdEncoding.DecodeString(sourceHref[0])

		if err != nil {
			w.WriteHeader(422)
			fmt.Fprintln(w, "href is not format")
			return
		}

		i := 0

		var resHeader *http.Response

		for resHeader, err = http.Head(string(headHref)); err != nil; i++ {
			if i >= 3 {
				//超过三次，失败
				w.WriteHeader(504)
				fmt.Fprintln(w, "can't find file head")
				return
			}
			time.Sleep(time.Microsecond * 300)
		}

		if len(resHeader.Header["Content-Length"]) <= 0 || len(resHeader.Header["Content-Length"][0]) <= 0 {
			size = "104857600" //100M 假设文件大小是100M
		} else {
			size = resHeader.Header["Content-Length"][0]
		}

	} else {
		_, err := strconv.Atoi(sourceSize[0])

		if err != nil {
			w.WriteHeader(422)
			fmt.Fprintln(w, "size is not formart")
			return
		}

		size = sourceSize[0]
	}

	sourceTitle := r.Form["title"]

	if len(sourceTitle) <= 0 || len(sourceTitle[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "title can't is null")
		return
	}

	id := strconv.Itoa(int(time.Now().UnixNano()))

	downloadPath, err := config.Config("downloadFilePath")

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "downloadFilePath is error")
		return
	}

	statusJson, _ := json.Marshal(map[string]string{"status": "waiting", "size": size, "href": downloadPath.(string) + "/" + id + "." + sourceType[0]})

	err = drive.RedisSetKeyValue(id, statusJson, 0)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "redis set ID value is fail")
		fmt.Fprintln(w, err)
		return
	}

	workJson, _ := json.Marshal(map[string]string{"id": id, "filename": sourceTitle[0], "size": size, "href": sourceHref[0], "type": sourceType[0], "fail": "0"})

	err = drive.RedisLPush("downloadList", workJson)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "redis set workList value is fail")
		fmt.Fprintln(w, err)
		return
	}

	downloadBaseHref, err := config.Config("downloadBaseHref")

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "config is error")
		fmt.Fprintln(w, err)
		return
	}

	resJson, _ := json.Marshal(map[string]string{"id": id, "url": downloadBaseHref.(string) + "/get_progress?id=" + id})

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(200)

	w.Write(resJson)
}

func GetTransformation(w http.ResponseWriter, r *http.Request) {

	//1.校验数据

	if r.Method != http.MethodGet {
		w.WriteHeader(404)
		fmt.Fprintln(w, "request method is not support")
		return
	}

	r.ParseForm() //解析参数

	id := r.Form["id"]

	if len(id) <= 0 || len(id[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "id can't is null")
		return
	}

	//2 获取当前状态和进度

	info, err := drive.RedisGetKeyValue(id[0])

	infoMap := map[string]string{}

	if err != nil {
		w.WriteHeader(422)
		fmt.Fprintln(w, "id not found or this job is timeout")
		return
	}

	err = json.Unmarshal([]byte(info), &infoMap)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "data to json error")
		return
	}

	switch infoMap["status"] {
	case "waiting":
		//计算下载进度 占50%

		//1.判断文件是否存在

		fileExists, err := fileUtil.PathExists(infoMap["href"])

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "system error")
			return
		}

		if !fileExists {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(200)

			resJson, _ := json.Marshal(map[string]float64{"percent": 0.00})

			w.Write(resJson)
			return
		}

		//2.计算下载进度

		size, err := strconv.Atoi(infoMap["size"]) //总大小

		if err != nil {
			//将此次转换设置为失败
			statusJson, _ := json.Marshal(map[string]string{"status": "fail"})

			drive.RedisSetKeyValue(id[0], statusJson, 0)
		}

		fileInfo, err := os.Stat(infoMap["href"])

		if err != nil {
			//将此次转换设置为失败

			statusJson, _ := json.Marshal(map[string]string{"status": "fail"})

			drive.RedisSetKeyValue(id[0], statusJson, 0)

		}

		if int(fileInfo.Size()) >= size {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(200)

			resJson, _ := json.Marshal(map[string]float64{"percent": 0.50})

			w.Write(resJson)
			return
		} else {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(200)

			value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(fileInfo.Size())/float64(size)/2.0), 64)

			resJson, _ := json.Marshal(map[string]float64{"percent": value})

			w.Write(resJson)
			return
		}

		break;
	case "download":
		//计算转换进度 占50%
		//1.判断文件是否存在

		fileExists, err := fileUtil.PathExists(infoMap["href"])

		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, "system error")
			return
		}

		if !fileExists {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(200)

			resJson, _ := json.Marshal(map[string]float64{"percent": 0.50})

			w.Write(resJson)
			return
		}

		value, err := tools.GetJobProgress(infoMap["href"])

		if err != nil {
			w.WriteHeader(463)
			fmt.Fprintln(w, "work is fail")

			return
		} else {

			if value >= 1 {
				value = 0.99
			}

			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(200)

			resJson, _ := json.Marshal(map[string]float64{"percent": value})

			w.Write(resJson)
			return
		}

		break;
	case "success":

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(200)

		resJson, _ := json.Marshal(map[string]interface{}{"percent": 1.00, "url": infoMap["href"]})

		w.Write(resJson)
		return
		break;
	case "fail":

		w.WriteHeader(463)
		fmt.Fprintln(w, "work is fail")

		return
		break;

	}

}
