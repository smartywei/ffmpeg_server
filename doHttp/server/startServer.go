package server

import (
	"net/http"
	"log"
	"ffmpeg/doHttp/controllers"
	"ffmpeg/doHttp/tools/config"
	"strconv"
	"fmt"
)

func StatHttpServer() {

	var serverPort string

	port, err := config.Config("port") //根据配置开启相应数目的协程

	if err != nil {
		panic(err)
	}

	_, ok := port.(float64)

	if ok {
		serverPort = strconv.Itoa(int(port.(float64)))
	} else {
		serverPort = "47017"
	}

	fmt.Println("http服务正在监听" + serverPort + "端口")

	http.HandleFunc("/transformation", controllers.StartTransformation) // 设置访问的路由
	http.HandleFunc("/get_progress", controllers.GetTransformation)     // 设置访问的路由
	http.HandleFunc("/download", controllers.DownloadFile)              // 设置访问的路由
	http.HandleFunc("/download_file", controllers.DownloadFileCache)        // 设置访问的路由
	http.HandleFunc("/download_file_local", controllers.DownloadFileToLocal)       // 设置访问的路由

	err = http.ListenAndServe(":"+serverPort, nil) // 设置监听的端口

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
