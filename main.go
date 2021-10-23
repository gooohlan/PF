package main

import (
	"flag"
	"fmt"

	"PF/engine"
	"PF/pf/parser"
)

// 通用性，不止能爬扑飞
// 以下流程写成通用的
// 1.获取漫画列表
// 2.漫画详情页获取图片链接
// 3.下载图片


var id = flag.String("i", "34802", "漫画id")

func main() {
	flag.Parse()
	engine.Run(engine.Request{
		Url:        fmt.Sprintf("%s/%s/%s/index.html", engine.PFUrl, "manhua", *id) ,
		Parser: engine.NewFuncParser(parser.ParseChapterList),
	})

}
