package server

import (
	"time"
	"ffmpeg/drive"
	"os"
	"net/http"
	"io"
	"encoding/json"
	"encoding/base64"
	"strconv"
	"ffmpeg/doHttp/tools"
	"ffmpeg/doHttp/tools/config"
	"fmt"
	"ffmpeg/doHttp/tools/fileUtil"
)

func StartDownloadServer() {

	var thead int

	downloadThead, err := config.Config("downloadThreadNum") //根据配置开启相应数目的协程

	if err != nil {
		thead = 1
	} else {
		_, ok := downloadThead.(float64)

		if ok {
			thead = int(downloadThead.(float64))
		} else {
			thead = 1
		}
	}

	if thead <= 0 {
		thead = 1
	}

	for i := 0; i < thead; i++ {
		go doDownloadServer()
	}

	fmt.Println(strconv.Itoa(thead) + "个下载进程已启动....")

}

//500毫秒一次去redis拿队列数据
func doDownloadServer() {

	for {
		ticker := time.NewTicker(time.Millisecond * 500)
		<-ticker.C
		res, err := drive.RedisRPop("downloadList")

		if err != nil {
			continue
		}
		info := map[string]string{}
		json.Unmarshal([]byte(res), &info)
		doDownload(info)
	}

}

func doDownload(info map[string]string) {

	id := info["id"]

	downloadPath, err := config.Config("downloadFilePath")

	if err != nil {
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，读取配置 downloadFilePath 失败.\t"+err.Error()+"\n")
		return
	}

	filePath := downloadPath.(string) + "/" + id + "." + info["type"]

	fileExist, err := fileUtil.PathExists(filePath)

	if err != nil {
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，判断 filePath 失败.\t"+err.Error()+"\n")
		return
	}


	if fileExist {

		err := os.Remove(filePath)

		if err != nil {
			tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，删除重复文件失败.\t"+err.Error()+"\n")
			return
		}
	}

	//判断任务重试次数
	fail, _ := strconv.Atoi(info["fail"])

	if fail >= 3 { //超过三次，设置为失败
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，重试三次仍然失败 .\t错误原因："+info["err"]+"\n")
		return
	}

	//需要下载文件
	f, err := os.Create(filePath)
	defer f.Close()

	if err != nil {
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，创建文件失败 .\t"+err.Error()+"\n")
		return
	}

	//进行base64 解码
	href, err := base64.StdEncoding.DecodeString(info["href"])

	if err != nil {
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，base64 解码href失败.\t"+err.Error()+"\n")
		return
	}

	resquest, err := http.Get(string(href))

	if err != nil || resquest.StatusCode != 200 { //重试队列

		if err != nil && fail == 2 {
			info["err"] = err.Error()
		}

		if err == nil && resquest.StatusCode != 200 {
			info["err"] = "请求下载，被拒绝！"
		}

		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("downloadList", resJson)
		return
	}

	_, err = io.Copy(f, resquest.Body)
	defer resquest.Body.Close()

	if err != nil { //重试队列
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("downloadList", resJson)
		return
	}

	//成功，改变redis 状态，添加转换队列
	targetPath, err := config.Config("outputFilePath")

	if err != nil {
		tools.SetJobFail(id, "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"时，读取配置 outputFilePath 失败 .\t"+err.Error()+"\n")
		return
	}

	statusJson, _ := json.Marshal(map[string]string{"status": "download", "href": targetPath.(string) + "/" + id + ".txt"})

	err = drive.RedisSetKeyValue(id, statusJson, 0)

	if err != nil { //重试队列
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("downloadList", resJson)
		return
	}

	jobJson, _ := json.Marshal(map[string]string{"id": id, "filename": info["filename"], "source": filePath, "target": targetPath.(string) + "/" + id + ".mp3", "log": targetPath.(string) + "/" + id + ".txt", "fail": "0"})

	err = drive.RedisLPush("jobList", jobJson)

	if err != nil { //重试队列
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("downloadList", resJson)
		return
	}

	tools.WriteLog("download", "["+tools.GetTimeString()+"]\t"+"下载id:"+id+"成功!"+"\n")
}
