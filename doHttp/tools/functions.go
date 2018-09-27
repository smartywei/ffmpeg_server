package tools

import (
	"os"
	"strconv"
	"regexp"
	"fmt"
	"encoding/json"
	"ffmpeg/drive"
	"ffmpeg/doHttp/tools/config"
	"ffmpeg/doHttp/tools/fileUtil"
	"time"
)

func GetJobIntTime(strTime string) int {

	h, _ := strconv.Atoi(string([]byte(strTime)[:2]))
	m, _ := strconv.Atoi(string([]byte(strTime)[3:5]))
	s, _ := strconv.Atoi(string([]byte(strTime)[6:8]))

	return h*3600 + m*60 + s
}

func GetJobProgress(filePath string) (float64, error) {

	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		return 0.0, err
	}

	info, err := os.Stat(filePath)

	if err != nil {
		return 0.0, err
	}

	res := make([]byte, info.Size())

	_, err = file.Read(res)

	if err != nil {
		return 0.0, err
	}

	r, _ := regexp.Compile("Duration: (.*?),")
	r2, _ := regexp.Compile("time=(.*?) ")

	time := r.Find(res)

	if len(time) <= 0 {
		return 0.50, err
	}

	countTime := GetJobIntTime(string(r.Find(res)[10:18]))

	time2 := r2.FindAll(res, -1)

	if len(time2) <= 0 {
		return 0.50, err
	}

	nowTime := GetJobIntTime(string(r2.FindAll(res, -1)[len(r2.FindAll(res, -1))-1][5:13]))

	value, err := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(nowTime)/float64(countTime)/2.0+0.50), 64)

	if err != nil {
		return 0.0, err
	}

	return value, nil
}

func SetJobFail(key string, content string) {
	statusJson, _ := json.Marshal(map[string]string{"status": "fail"})
	err := drive.RedisSetKeyValue(key, statusJson, 7200*time.Second)
	if err != nil {
		WriteLog("error", "["+GetTimeString()+"]\t"+"设置 redis 失败.\t"+err.Error()+"\n")
	}
	WriteLog("error", content)
}

func GetYearMonthString() string {
	y, m, _ := time.Now().Date()
	return strconv.Itoa(y) + "-" + strconv.Itoa(int(m))
}

func GetDateString() string {
	y, m, d := time.Now().Date()
	return strconv.Itoa(y) + "-" + strconv.Itoa(int(m)) + "-" + strconv.Itoa(d)
}

func GetTimeString() string {
	h := time.Now().Hour()
	m := time.Now().Minute()
	s := time.Now().Second()

	return GetDateString() + " " + strconv.Itoa(h) + ":" + strconv.Itoa(m) + ":" + strconv.Itoa(s)
}

func WriteLog(mode string, content string) {

	logPath, err := config.Config("logPath")

	if err != nil {
		return
	}

	switch mode {
	case "error":
		file, err := fileUtil.CreateFile(logPath.(string)+"/error/"+GetYearMonthString(), GetDateString()+".log")
		defer file.Close()

		if err != nil {
			return
		}

		fileUtil.WriteFile(file, content)
		break;
	case "download":
		file, err := fileUtil.CreateFile(logPath.(string)+"/download/"+GetYearMonthString(), GetDateString()+".log")
		defer file.Close()

		if err != nil {
			return
		}

		fileUtil.WriteFile(file, content)
		break;
	case "job":
		file, err := fileUtil.CreateFile(logPath.(string)+"/job/"+GetYearMonthString(), GetDateString()+".log")
		defer file.Close()

		if err != nil {
			return
		}

		fileUtil.WriteFile(file, content)
		break;
	default:
		return
	}
}