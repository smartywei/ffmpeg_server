package controllers

import (
	"net/http"
	"os"
	"io"
	"fmt"
	"ffmpeg/drive"
	"encoding/json"
	"strconv"
	"strings"
	"net"
	"encoding/base64"
	"ffmpeg/doHttp/tools/config"
	"time"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {

	r.ParseForm() //解析参数

	id := r.Form["id"]

	if len(id) <= 0 || len(id[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "id can't is null")
		return
	}

	info, err := drive.RedisGetKeyValue(id[0])

	if err != nil {
		w.WriteHeader(404)
		return
	}

	downloadInfo := map[string]string{}

	json.Unmarshal([]byte(info), &downloadInfo)

	fileFullPath := downloadInfo["source"]

	file, err := os.Open(fileFullPath)
	defer file.Close()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "file not found")
		return
	}

	stat, err := file.Stat()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "file not get stat")
		return
	}

	fileName := downloadInfo["filename"] // 防止乱码 ??

	w.Header().Set("Content-Length", strconv.Itoa(int(stat.Size())))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	_, error := io.Copy(w, file)
	if error != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "download error,please try again")
		return
	}
}

func DownloadFileCache(w http.ResponseWriter, r *http.Request) {

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

	sourceTitle := r.Form["title"]

	if len(sourceTitle) <= 0 || len(sourceTitle[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "title can't is null")
		return
	}

	sourceType := r.Form["type"]

	if len(sourceType) <= 0 || len(sourceType[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "type can't is null")
		return
	}

	headHref, err := base64.StdEncoding.DecodeString(sourceHref[0])

	var i = 0
	var resHeader *http.Response
	var size string

	for resHeader, err = http.Head(string(headHref)); err != nil; i++ {
		if i >= 3 {
			//超过三次，失败
			w.WriteHeader(504)
			fmt.Fprintln(w, "can't get file")
			return
		}
		time.Sleep(time.Microsecond * 250)
	}

	if len(resHeader.Header["Content-Length"]) <= 0 || len(resHeader.Header["Content-Length"][0]) <= 0 {
		size = ""
	} else {
		size = resHeader.Header["Content-Length"][0]
	}

	id := strconv.Itoa(int(time.Now().UnixNano()))

	statusJson, _ := json.Marshal(map[string]string{"href": sourceHref[0], "title": sourceTitle[0] + "." + sourceType[0], "size": size})

	err = drive.RedisSetKeyValue(id, statusJson, 7200*time.Second)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "redis set ID value is fail")
		fmt.Fprintln(w, err)
		return
	}

	downloadBaseHref, err := config.Config("downloadBaseHref")

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "read config is fail")
		fmt.Fprintln(w, err)
		return
		return
	}

	resJson, _ := json.Marshal(map[string]string{"url": downloadBaseHref.(string) + "/download_file_local?id=" + id})

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(200)

	w.Write(resJson)

}

func DownloadFileToLocal(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //解析参数

	id := r.Form["id"]

	if len(id) <= 0 || len(id[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "id can't is null")
		return
	}

	info, err := drive.RedisGetKeyValue(id[0])

	if err != nil {
		w.WriteHeader(404)
		return
	}

	downloadInfo := map[string]string{}

	json.Unmarshal([]byte(info), &downloadInfo)

	href, _ := base64.StdEncoding.DecodeString(downloadInfo["href"])

	fileName := downloadInfo["title"] // 防止中文乱码

	resquest, err := http.Get(string(href))

	if err != nil || resquest.StatusCode != 200 { //下载失败
		w.WriteHeader(403)
		return
	}

	if downloadInfo["size"] != "" {
		w.Header().Set("Content-Length", downloadInfo["size"])
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	_, err = io.Copy(w, resquest.Body)
	defer resquest.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		return
	}

}
