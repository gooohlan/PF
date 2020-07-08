package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
	"github.com/gocolly/colly"
	"github.com/robertkrimen/otto"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	ImgHeader = "http://res.img.fffmanhua.com/"
	PF        = "http://www.pufei8.com"
	MANHUA    = PF + "/manhua/"
	DIR       = "./PF/"
)

// 访问漫画首页,获取目录
// 通过目录链接获取该目录下所有图片地址,
// 通过图片地址,获取所有的图片
// 存入对应文件夹

// 漫画目录
type Catalog struct {
	Title  string
	Url    string
	ImgArr []string
}

var catalog []*Catalog
var MaxNum int

func main() {
	//if len(os.Args) < 2 {
	//	log.Println("请输入Id")
	//}
	//Mid, err := strconv.Atoi(os.Args[1])
	//if err != nil {
	//	log.Println("漫画ID格式错误")
	//	panic(err)
	//}
	Mid := 419
	// 目标漫画对应Id
	c := colly.NewCollector()
	c.OnHTML("#play_0", func(e *colly.HTMLElement) {
		e.ForEach("ul li a", func(i int, element *colly.HTMLElement) {
			href := element.Attr("href")
			title := element.Text
			//title = iconv.ConvertString(title, "GB2312", "utf-8")
			title = coverGBKToUTF8(title)
			catalog = append(catalog, &Catalog{Url: PF + href, Title: title})
		})
	})

	c.OnRequest(func(r *colly.Request) {
	})

	c.Visit(MANHUA + strconv.Itoa(Mid))

	var wg sync.WaitGroup

	for k, v := range catalog {
		wg.Add(1)
		go func(v *Catalog, k int) {
			//获取图片地址
			GetImgArr(v)
			if runtime.NumGoroutine() > MaxNum {
				MaxNum = runtime.NumGoroutine()
			}
			//创建文件夹并获取图片
			CreateFileGetImg(v, len(catalog)-k)
			wg.Done()
		}(v, k)
	}
	time.Sleep(time.Second)
	log.Println(MaxNum)
}

// 获取该章节所有图片地址
func GetImgArr(catalog *Catalog) {
GETURL:
	resp, err := http.Get(catalog.Url)
	if err != nil {
		log.Println("获取漫画子页面失败1:", err)
		goto GETURL
	}
	bodyReader := bufio.NewReader(resp.Body)
	e := determineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	result, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		log.Println("解析页面失败", err)
		//panic(err)
	}
	resp.Body.Close()
	var r = regexp.MustCompile("fu[\\S\\W]+};")
	JavaScript := r.FindString(string(result))
	JavaScript += " function f() {return photosr;} f();"
	vm := otto.New()
	value, err := vm.Run(JavaScript)
	if err != nil {
		log.Println("解析图片地址失败:", err)
		//panic(err)
	}
	imgStr, err := value.ToString()
	if err != nil {
		log.Println("图片地址解析出错:", err)
		//panic(err)
	}
	imgArr := strings.Split(imgStr, ",")
	catalog.ImgArr = imgArr[1:]

}

//编码统一为utf8
func determineEncoding(r *bufio.Reader) (e encoding.Encoding) {
	bytes, err := r.Peek(1024)
	if err != nil {
		return unicode.UTF8
	}
	e, _, _ = charset.DetermineEncoding(bytes, "")
	return
}

// 创建文件夹并存储图片
func CreateFileGetImg(catalog *Catalog, index int) {
	if catalog.Title == "通知" {
		return
	}
	dir := DIR + strconv.Itoa(index) + "--" + catalog.Title
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Println("创建文件夹失败")
	}
	var wg sync.WaitGroup
	for k, v := range catalog.ImgArr {
		wg.Add(1)
		go func(img string, k int) {
			if runtime.NumGoroutine() > MaxNum {
				MaxNum = runtime.NumGoroutine()
			}
		GETIMAGE:
			resp, err := http.Get(ImgHeader + img)
			if err != nil {
				log.Println(err)
				goto GETIMAGE
			}
			body, _ := ioutil.ReadAll(resp.Body)
			out, _ := os.Create(dir + "/" + strconv.Itoa(k+1) + ".jpg")
			io.Copy(out, bytes.NewReader(body))
			resp.Body.Close()
			wg.Done()
		}(v, k)
	}
	wg.Wait()
}

func coverGBKToUTF8(src string) string {
	// 网上搜有说要调用translate函数的，实测不用
	return mahonia.NewDecoder("gbk").ConvertString(src)
}
