package parser

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"PF/common"
	"PF/engine"

	"github.com/robertkrimen/otto"
)

var reComicList = regexp.MustCompile(`<script\b[^>]*>([\s\S]*?)</script>`)
var reHead = regexp.MustCompile(`<head>[\s\S]+</head>`)

func ParseComicList(contents []byte, title string) engine.ParseResult {
	log.Println("create comicList title:", title)
	head := reHead.Find(contents)
	matches := reComicList.FindAllSubmatch(head, 1)
	result := engine.ParseResult{}
	if len(matches) != 1 {
		return result
	}
	scriptStr := string(matches[0][1])
	scriptStr += " function f() {return photosr;} f();"
	vm := otto.New()
	value, err := vm.Run(scriptStr)
	if err != nil {
		log.Println("解析图片地址失败:", err)
		return result
	}
	comicListStr, err := value.ToString()
	if err != nil {
		log.Println("解析图片地址失败:", err)
		return result
	}
	comicList := strings.Split(comicListStr, ",")
	if len(comicList) <= 2 {
		log.Println("图片获取失败..............")
		return result
	}
	comicList = comicList[1:]
	for i, comic := range comicList {
		result.Items = append(result.Items, "Image List "+title+"-"+strconv.Itoa(i+1))
		result.Requests = append(result.Requests, engine.Request{
			Url:    common.UrlJoin(engine.ImgUrl, comic),
			Type: engine.Image,
			Parser: NewImageParser(title + "-" + strconv.Itoa(i+1)),
		})
	}
	return result
}

type ComicListParser struct {
	name string
}

func NewComicListParser(name string) *ComicListParser {
	return &ComicListParser{name: name}
}

func (c ComicListParser) Parse(contents []byte) engine.ParseResult {
	return ParseComicList(contents, c.name)
}
