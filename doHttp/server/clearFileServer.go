package server

import (
	"time"
	"ffmpeg/doHttp/tools"
	"ffmpeg/doHttp/tools/config"
	"os"
	"io/ioutil"
	"ffmpeg/doHttp/tools/fileUtil"
	"fmt"
)

func StartClearFileServer() {

	var clearTime time.Duration
	var outTime int64

	//获取文件过期时间
	fileOutTime, err := config.Config("clearFileTime")

	if err != nil {
		outTime = 3 * 60 * 60
	} else {
		_, ok := fileOutTime.(float64)

		if ok {
			outTime = int64(fileOutTime.(float64))
		} else {
			outTime = 3 * 60 * 60
		}
	}

	// 根据配置设置文件清理时间周期
	clearFileTime, err := config.Config("clearFileTime")

	if err != nil {
		clearTime = 12 * 60 * 60
	} else {
		_, ok := clearFileTime.(float64)

		if ok {
			clearTime = time.Duration(clearFileTime.(float64))
		} else {
			clearTime = 12 * 60 * 60
		}
	}

	fmt.Println("文件清理服务，已启动...")

	ticker := time.NewTicker(time.Second * clearTime)

	for _ = range ticker.C {
		doClearFile(outTime)
	}

}

func doClearFile(clearTime int64) {
	// 清理过期文件（超过配置文件时间的文件，音视频和日志）

	//获取需要清理的文件目录

	//读取下载目录
	downloadFilePath, err := config.Config("downloadFilePath")

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取下载目录配置失败！")
		return
	}

	downloadPath := downloadFilePath.(string)

	fileExist, err := fileUtil.PathExists(downloadPath)

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取 downloadPath 失败！")
		return
	}

	if !fileExist {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，配置的下载目录不存在！")
		return
	}

	//读取输出目录
	outputFilePath, err := config.Config("outputFilePath")

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取输出目录配置失败！")
		return
	}

	outputPath := outputFilePath.(string)

	fileExist, err = fileUtil.PathExists(outputPath)

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取 outputPath 失败！")
		return
	}

	if !fileExist {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，配置的输出目录不存在！")
		return
	}

	timeOut := time.Now().UnixNano() - (clearTime * 1000 * 1000 * 1000)

	//开始清理下载文件
	downloadFiles, err := ioutil.ReadDir(downloadPath)

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取下载目录文件失败！")
		return
	}

	for _, file := range downloadFiles {
		if file.ModTime().UnixNano() < timeOut {
			err := os.Remove(downloadPath + "/" + file.Name())

			if err != nil {
				tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理下载文件"+file.Name()+"失败！")
			}

		}
	}

	//开始清理输出文件
	ouputFiles, err := ioutil.ReadDir(outputPath)

	if err != nil {
		tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理文件时，读取输出目录文件失败！")
		return
	}

	for _, file := range ouputFiles {
		if file.ModTime().UnixNano() < timeOut {
			err := os.Remove(outputPath + "/" + file.Name())

			if err != nil {
				tools.WriteLog("error", "["+tools.GetTimeString()+"]\t"+"清理输出文件"+file.Name()+"失败！")
			}

		}
	}

}
