package main

import (
	"ffmpeg/doHttp/server"
	"fmt"
	"ffmpeg/doHttp/tools/config"
)

func main() {

	//检查环境，加载配置等

	//1.检查配置文件是否存在，如果不存在，则生成默认配置文件
	err := config.CreateConfigFile()

	if err != nil {
		panic(err)
	}

	//2.根据配置检查其他文件目录是否存在，如果不存在，则生成

	err = config.CreateOtherDir()

	if err != nil {
		panic(err)
	}

	start := make(chan bool,1)

	fmt.Println("HTTP 服务已启动...")

	go server.StatHttpServer() //启动http服务

	go server.StartDownloadServer() //启动下载队列监听

	go server.StartJobServer()//启动转码队列监听

	go server.StartClearFileServer() //启动清理文件队列监听

	<-start
}
