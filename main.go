package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/baidu-tools/util"
	"io/ioutil"
	"log"
)

var (
	filename = flag.String("f", "", "file name.")
)

func main() {
	flag.Parse()
	baiduUtil.SetLogPrefix()

	if *filename == "" {
		flag.Usage()
		return
	}

	log.Println("正在上传图片, 请稍后")
	data, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatalln(err)
	}

	imgJSON, err := baiduUtil.Fetch("POST", "https://tinypng.com/web/shrink", nil, data, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		log.Fatalln(err)
	}

	json, err := simplejson.NewJson(imgJSON)
	if err != nil {
		log.Fatalln(err)
	}

	if j, ok := json.CheckGet("error"); ok {
		log.Fatalln(j.MustString(), ":", json.Get("message").MustString())
	}

	outputJSON := json.Get("output")
	url := outputJSON.Get("url").MustString()

	log.Println("上传图片成功, 正在下载压缩后的图片...")
	img, err := baiduUtil.Fetch("GET", url, nil, nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(*filename, img, 0666)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("图片保存成功")
	log.Printf("图片类型: %s, 原始图片大小: %d, 压缩后图片大小: %d, 压缩比率: %f\n", outputJSON.Get("type").MustString(), json.GetPath("input", "size").MustInt(), outputJSON.Get("size").MustInt(), outputJSON.Get("ratio").MustFloat64())
}
