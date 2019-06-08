package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"github.com/iikira/BaiduPCS-Go/requester/uploader"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// Version 版本号
	Version = "v1.0"
)

var (
	isOverWrite  = false
	printVersion = false
	outputDir    string
)

func init() {
	flag.BoolVar(&isOverWrite, "w", false, "overwrite")
	flag.BoolVar(&printVersion, "v", false, "print version")
	flag.StringVar(&outputDir, "o", ".", "output directory")

	flag.Parse()

	if printVersion {
		fmt.Printf("TinyPNG client, Version %s\n", Version)
		fmt.Println("Copyright (c) 2017-2018, iikira/tinypng: https://github.com/iikira/tinypng")
		os.Exit(0)
	}
	pcsutil.SetLogPrefix()
}

func main() {
	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	for k := range flag.Args() {
	for_1:
		for errTimes := 0; errTimes < 3; errTimes++ {
			code := do(flag.Arg(k))

			switch code {
			case 0:
				fallthrough
			case 1: // 系统错误
				break for_1
			case 2: // 网络错误
				fallthrough
			case 3: // json 解析错误
				continue
			}
		}
	}
}

type UploadedData struct {
	Input   InputData  `json:"input"`
	Output  OutputData `json:"output"`
	Error   string     `json:"error"`
	Message string     `json:"message"`
}

type InputData struct {
	Size int64  `json:"size"`
	Type string `json:"type"`
}

type OutputData struct {
	InputData
	Width  uint    `json:"width"`
	Height uint    `json:"height"`
	Ratio  float64 `json:"ratio"`
	URL    string  `json:"url"`
}

func do(filename string) (code int) {
	var savePath string
	if isOverWrite {
		savePath = filename
	} else {
		savePath = filepath.Clean(outputDir + string(os.PathSeparator) + filepath.Dir(filename) + string(os.PathSeparator) + "tinified-" + filepath.Base(filename))
		_, err := os.Stat(savePath)
		if err == nil { // 文件已存在
			log.Printf("[%s] 已存在\n", savePath)
			return 1
		}
	}

	log.Printf("[%s] 正在上传图片\n", filename)
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		log.Println(err)
		return 1
	}

	// 保留权限
	mode := os.FileMode(0666)
	info, err := file.Stat()
	if err == nil {
		mode = info.Mode()
	}

	uploader.DoUpload("https://tinypng.com/web/shrink", rio.NewFileReaderLen64(file), func(resp *http.Response, err error) {
		file.Close()
		fmt.Println()

		if err != nil {
			log.Println(err)
			code = 2
			return
		}

		imgJSON, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			code = 1
			return
		}

		resp.Body.Close()

		data := new(UploadedData)

		err = json.Unmarshal(imgJSON, data)
		if err != nil {
			log.Println(err)
			code = 3
			return
		}

		if data.Error != "" {
			log.Printf("[%s] Error, %s: %s\n", filename, data.Error, data.Message)
			code = 1
			return
		}

		log.Printf("[%s] 上传图片成功, 正在下载压缩后的图片...\n", filename)

		img, err := requester.Fetch("GET", data.Output.URL, nil, nil)
		if err != nil {
			log.Println(err)
			code = 2
			return
		}

		imgLen := int64(len(img))
		if imgLen != data.Output.Size {
			log.Printf("[%s] 图片下载失败, 文件大小不一致, 已下载: %d, 远程: %d\n", filename, imgLen, data.Output.Size)
			code = 2
			return
		}

		err = ioutil.WriteFile(savePath, img, mode)
		if err != nil {
			log.Println(err)
			code = 1
			return
		}

		log.Printf("[%s] 图片保存成功, 保存位置: %s, 图片类型: %s, 图片宽度: %d, 图片高度: %d, 原始图片大小: %s, 压缩后图片大小: %s, 压缩比率: %f%%\n", filename, savePath, data.Output.Type, data.Output.Width, data.Output.Height, converter.ConvertFileSize(data.Input.Size), converter.ConvertFileSize(data.Output.Size), data.Output.Ratio*100)
	})

	return code
}
