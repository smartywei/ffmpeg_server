package config

import (
	"os"
	"encoding/json"
	"errors"
	"ffmpeg/doHttp/tools/fileUtil"
)

func CreateConfigFile() error {

	confIsExist, err := fileUtil.PathExists("./config/server.conf.json")

	if err != nil{
		return err
	}

	if !confIsExist {

		err := fileUtil.CreateDir("./config")

		if err != nil {
			return err
		}

		f, err := os.Create("./config/server.conf.json")

		defer f.Close()

		if err != nil {
			return err
		}

		// 写入默认配置
		defaultConfigMap := map[string]interface{}{
			"downloadThreadNum": 1,
			"jobThreadNum":      1,
			"port":              47017,
			"downloadBaseHref":  "127.0.0.1:47017",
			"downloadFilePath":  "./download",
			"outputFilePath":    "./output",
			"logPath":           "./log",
			"fileTimeOut":       3 * 60 * 60,  //转换文件下载过期时间
			"clearFileTime":     12 * 60 * 60, // 默认12个小时清理一次
			"transformationIPs": "127.0.0.1",  // 可以访问转换的IP
			"redisHost":         "127.0.0.1",
			"redisPort":         "6379",
			"redisPassword":     "",
		}

		defaultConfigJson, err := json.MarshalIndent(defaultConfigMap, "", "	")

		if err != nil {
			return err
		}

		_, err = f.Write([]byte(defaultConfigJson))

		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func Config(key string) (interface{}, error) {

	if key != "downloadBaseHref" &&
		key != "downloadFilePath" &&
		key != "downloadThreadNum" &&
		key != "jobThreadNum" &&
		key != "logPath" &&
		key != "outputFilePath" &&
		key != "fileTimeOut" &&
		key != "clearFileTime" &&
		key != "transformationIPs" &&
		key != "redisHost" &&
		key != "redisPort" &&
		key != "redisPassword" &&
		key != "port" {

		return "", errors.New("key is not find")
	}

	confExist, err := fileUtil.PathExists("./config/server.conf.json")

	if err != nil {
		return nil,err
	}

	if !confExist {
		CreateConfigFile()
	}

	f, err := os.Open("./config/server.conf.json")
	defer f.Close()

	if err != nil {
		return "", err
	}

	stat, err := f.Stat()

	if err != nil {
		return "", err
	}

	fileBody := make([]byte, stat.Size())

	_, err = f.Read(fileBody)

	if err != nil {
		return "", err
	}

	res := map[string]interface{}{}

	err = json.Unmarshal(fileBody, &res)

	if err != nil {
		return "", err
	}

	if _, ok := res[key]; !ok {
		switch key {
		case "downloadBaseHref":
			return "127.0.0.1:47017", nil
		case "downloadFilePath":
			return "./download", nil
		case "downloadThreadNum":
			return 1, nil
		case "jobThreadNum":
			return 1, nil
		case "logPath":
			return "./log", nil
		case "outputFilePath":
			return "./output", nil
		case "port":
			return 47017, nil
		case "fileTimeOut":
			return 3 * 60 * 60, nil
		case "clearFileTime":
			return 12 * 60 * 60, nil
		case "transformationIPs":
			return "127.0.0.1", nil
		case "redisHost":
			return "127.0.0.1", nil
		case "redisPassword":
			return "", nil
		case "redisPort":
			return "6379", nil

		}
	}

	return res[key], nil
}

func CreateOtherDir() error {

	downloadFilePath, err := Config("downloadFilePath")

	if err != nil {
		return err
	}

	outputFilePath, err := Config("outputFilePath")

	if err != nil {
		return err
	}

	logPath, err := Config("logPath")

	if err != nil {
		return err
	}

	err = fileUtil.CreateDir(downloadFilePath.(string))

	if err != nil {
		return err
	}

	err = fileUtil.CreateDir(outputFilePath.(string))

	if err != nil {
		return err
	}

	err = fileUtil.CreateDir(logPath.(string) + "/error")

	if err != nil {
		return err
	}

	err = fileUtil.CreateDir(logPath.(string) + "/download")

	if err != nil {
		return err
	}

	err = fileUtil.CreateDir(logPath.(string) + "/job")

	if err != nil {
		return err
	}

	return nil
}
