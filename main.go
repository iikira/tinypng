package main

import (
	"flag"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/baidu-tools/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	isOverWrite = flag.Bool("w", false, "over write mode")
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

func do(filename string) (code int) {
	log.Printf("[%s] 正在上传图片\n", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return 1
	}

	imgJSON, err := baiduUtil.Fetch("POST", "https://tinypng.com/web/shrink", nil, data, map[string]string{
		// "Content-Type": "application/x-www-form-urlencoded",
		"Content-Encoding": "gzip",
	})
	if err != nil {
		log.Println(err)
		return 2
	}

	json, err := simplejson.NewJson(imgJSON)
	if err != nil {
		log.Println(err)
		return 3
	}

	if j, ok := json.CheckGet("error"); ok {
		log.Printf("[%s] Error, %s: %s\n", filename, j.MustString(), json.Get("message").MustString())
		return 1
	}

	outputJSON := json.Get("output")
	url := outputJSON.Get("url").MustString()

	log.Printf("[%s] 上传图片成功, 正在下载压缩后的图片...\n", filename)
	img, err := baiduUtil.Fetch("GET", url, nil, nil, nil)
	if err != nil {
		log.Println(err)
		return 2
	}

	outputSize := outputJSON.Get("size").MustFloat64()
	if len(img) != int(outputSize) {
		log.Printf("[%s] 图片下载失败, 文件大小不一致\n", filename)
		return 2
	}

	var newName string
	if *isOverWrite {
		newName = filename
	} else {
		newName = filepath.Dir(filename) + "/tinified-" + filepath.Base(filename)
	}
	err = ioutil.WriteFile(newName, img, 0666)
	if err != nil {
		log.Println(err)
		return 1
	}
	log.Printf("[%s] 图片保存成功, 保存位置: %s, 图片类型: %s, 原始图片大小: %s, 压缩后图片大小: %s, 压缩比率: %f%%\n", filename, newName, outputJSON.Get("type").MustString(), convertSize(json.GetPath("input", "size").MustFloat64()), convertSize(outputSize), outputJSON.Get("ratio").MustFloat64()*100)
	return 0
}
