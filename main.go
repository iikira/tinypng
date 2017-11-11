package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/baidu-tools/util"
	"io/ioutil"
	"log"
	"os"
)

func init() {
	flag.Parse()
	baiduUtil.SetLogPrefix()
	baiduUtil.SetTimeout(3e11)
}

func main() {
	if len(os.Args) <= 1 {
		log.Println("请输入参数")
		flag.Usage()
		return
	}

	for k := range flag.Args() {
		do(flag.Arg(k))
	}

}

func do(filename string) {
	log.Printf("正在上传图片 %s\n", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return
	}

	imgJSON, err := baiduUtil.Fetch("POST", "https://tinypng.com/web/shrink", nil, data, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		log.Println(err)
		return
	}

	json, err := simplejson.NewJson(imgJSON)
	if err != nil {
		log.Println(err)
		return
	}

	if j, ok := json.CheckGet("error"); ok {
		log.Fatalln(j.MustString(), ":", json.Get("message").MustString())
	}

	outputJSON := json.Get("output")
	url := outputJSON.Get("url").MustString()

	log.Println("上传图片成功, 正在下载压缩后的图片...")
	img, err := baiduUtil.Fetch("GET", url, nil, nil, nil)
	if err != nil {
		log.Println(err)
		return
	}

	outputSize := outputJSON.Get("size").MustInt()
	if len(img) != outputSize {
		log.Println("图片下载失败, 文件大小不一致")
		return
	}

	err = ioutil.WriteFile(filename, img, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("图片保存成功")
	log.Printf("图片类型: %s, 原始图片大小: %d, 压缩后图片大小: %d, 压缩比率: %f\n", outputJSON.Get("type").MustString(), json.GetPath("input", "size").MustInt(), outputSize, outputJSON.Get("ratio").MustFloat64())
}
