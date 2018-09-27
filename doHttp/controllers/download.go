package controllers

import (
	"net/http"
	"os"
	"io"
	"fmt"
	"ffmpeg/drive"
	"encoding/json"
	"strconv"
)

func DownloadFile(w http.ResponseWriter, r *http.Request) {

	r.ParseForm() //解析参数

	id := r.Form["id"]

	if len(id) <= 0 || len(id[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "id can't is null")
		return
	}

	info,err := drive.RedisGetKeyValue(id[0])

	if err != nil{
		w.WriteHeader(404)
		return
	}

	downloadInfo := map[string]string{}

	json.Unmarshal([]byte(info),&downloadInfo)

	fileFullPath := downloadInfo["source"]

	file, err := os.Open(fileFullPath)
	defer file.Close()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "file not found")
		return
	}

	stat,err:= file.Stat()

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "file not get stat")
		return
	}

	fileName := downloadInfo["filename"] // 防止乱码 ??

	w.Header().Set("Content-Length",strconv.Itoa(int(stat.Size())))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

	_, error := io.Copy(w,file)
	if error != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "download error,please try again")
		return
	}
}
