package server

import (
	"time"
	"ffmpeg/drive"
	"encoding/json"
	"os"
	"strconv"
	"ffmpeg/doHttp/tools"
	"ffmpeg/doHttp/tools/config"
	"fmt"
)

func StartJobServer() {

	var thead int

	jobThead, err := config.Config("jobThreadNum") //根据配置开启相应数目的协程

	if err != nil {
		thead = 1
	} else {

		_, ok := jobThead.(float64)

		if ok {
			thead = int(jobThead.(float64))
		} else {
			thead = 1
		}
	}

	if thead <= 0 {
		thead = 1
	}

	for i := 0; i < thead; i++ {
		go doJobServer()
	}

	fmt.Println(strconv.Itoa(thead) + "个转换进程已启动....")
}

func doJobServer() {

	for {
		ticker := time.NewTicker(time.Millisecond * 500)
		<-ticker.C
		res, err := drive.RedisRPop("jobList")

		if err != nil {
			continue
		}

		info := map[string]string{}

		json.Unmarshal([]byte(res), &info)

		doJob(info)
	}

}

func doJob(info map[string]string) {

	//判断任务重试次数
	fail, err := strconv.Atoi(info["fail"])

	if err != nil {
		tools.SetJobFail(info["id"], "["+tools.GetTimeString()+"]\t"+"转换id:"+info["id"]+"时，获取失败次数失败./t"+err.Error()+"\n")
		return
	}

	if fail >= 3 { //超过三次，设置为失败
		tools.SetJobFail(info["id"], "["+tools.GetTimeString()+"]\t"+"转换id:"+info["id"]+"时，重试三次仍然失败,错误原因："+info["err"]+"\n")
		return
	}

	file, err := os.Create(info["log"])
	defer file.Close()

	if err != nil {
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("jobList", resJson)
		return
	}

	cmd := drive.ToDefaultTransformation(info["source"], info["target"])

	cmd.Stderr = file

	err = cmd.Run()

	if err != nil {
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("jobList", resJson)
		return
	}

	downloadBaseHref, err := config.Config("downloadBaseHref")

	if err != nil {
		tools.SetJobFail(info["id"], "["+tools.GetTimeString()+"]\t"+"转换id:"+info["id"]+"时，读取配置 downloadBaseHref 失败.\t"+err.Error()+"\n")
		return
	}

	statusJson, err := json.Marshal(map[string]string{"status": "success", "source": info["target"], "href": downloadBaseHref.(string) + "/download?id=" + info["id"], "filename": info["filename"] + ".mp3"})

	if err != nil {
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("jobList", resJson)
		return
	}

	//设置过期时间
	var timeout time.Duration

	fileTimeOut, err := config.Config("fileTimeOut")

	if err != nil {
		tools.SetJobFail(info["id"], "["+tools.GetTimeString()+"]\t"+"转换id:"+info["id"]+"时，读取配置 fileTimeOut 失败.\t"+err.Error()+"\n")
		return
	}

	_, ok := fileTimeOut.(float64)

	if ok {
		timeout = time.Duration(fileTimeOut.(float64))
	} else {
		timeout = 3 * 60 * 60
	}

	err = drive.RedisSetKeyValue(info["id"], statusJson, time.Second*timeout)

	if err != nil {
		if fail == 2 {
			info["err"] = err.Error()
		}
		info["fail"] = strconv.Itoa(fail + 1)
		resJson, _ := json.Marshal(info)
		drive.RedisRPush("jobList", resJson)
		return
	}

	tools.WriteLog("job", "["+tools.GetTimeString()+"]\t"+"转换id:"+info["id"]+"成功！"+"\n")
}
