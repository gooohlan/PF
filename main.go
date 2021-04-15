package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robertkrimen/otto"

	"github.com/axgle/mahonia"
	"github.com/gocolly/colly"
)

const (
	ImgHeader = "http://res.img.fffmanhua.com/"
	PF        = "http://www.pufei8.com"
	MANHUA    = PF + "/manhua/"
	DIR       = "./PF/"
)

// 访问漫画首页,获取目录
// 通过目录链接获取该目录下所有图片地址
// 通过图片地址,获取所有的图片,存入对应文件夹

// 漫画目录
type Catalog struct {
	Title  string
	Url    string
	ImgArr []string
}

var catalog []*Catalog // 漫画目录
var Title string       // 漫画标题

func main() {
	var Mid int
	fmt.Println("请输入漫画Id：")
	_, err := fmt.Scanln(&Mid)
	if err != nil {
		log.Println("漫画ID格式错误")
		panic(err)
	}

	c := colly.NewCollector()
	c.OnHTML("#play_0", func(e *colly.HTMLElement) {
		e.ForEach("ul li a", func(i int, element *colly.HTMLElement) {
			href := element.Attr("href")
			title := element.Text
			title = coverGBKToUTF8(title) // 页面编码转为UTF8
			catalog = append(catalog, &Catalog{Url: PF + href, Title: title})
		})
	})
	// 获取漫画标题,并创建目录
	c.OnHTML(".titleInfo h1", func(element *colly.HTMLElement) {
		Title = coverGBKToUTF8(element.Text)
		err := os.MkdirAll(Title, os.ModePerm)
		if err != nil {
			log.Println("创建文件夹失败")
		}
	})

	c.OnRequest(func(r *colly.Request) {
	})

	c.Visit(MANHUA + strconv.Itoa(Mid))

	pool := New(300)
	for k, v := range catalog {
		pool.Add(1)
		go func(v *Catalog, k int) {
			GetImgArr(v)
			pool.Done()
		}(v, k)
	}
	pool.Wait()
	time.Sleep(time.Second * 1)

	log.Println("开始下载图片.....")
	var wg sync.WaitGroup
	var jobsChan = make(chan int, 15)
	poolCount := 15
	for i := 0; i < poolCount; i++ {
		go func() {
			for j := range jobsChan {
				CreateFileGetImg(catalog[j], len(catalog)-j)
				wg.Done()
				time.Sleep(time.Microsecond * 500)
			}
		}()
	}
	var NullCatalog []*Catalog
	for i := 0; i < len(catalog); i++ {
		jobsChan <- i
		wg.Add(1)
		log.Println("开始下载章节:", catalog[i].Title, "共", len(catalog[i].ImgArr), "页")
		if len(catalog[i].ImgArr) == 0 {
			NullCatalog = append(NullCatalog, catalog[i])
			log.Println("图片目录为空..........")
		}
	}
	wg.Wait()
	time.Sleep(time.Second)
	log.Println("下载完成")
	if len(NullCatalog) > 0 {
		log.Println("以下章节下载失败,请前往扑飞漫画查看")
		for _, v := range NullCatalog {
			log.Println(v.Title, ":", v.Url)
		}
	}
}

// 获取该章节所有图片地址
func GetImgArr(catalog *Catalog) {
	c := colly.NewCollector()
	c.OnHTML("head script", func(e *colly.HTMLElement) {
		if e.Text != "" {
			JavaScript := coverGBKToUTF8(e.Text)
			JavaScript += " function f() {return photosr;} f();"
			vm := otto.New()
			value, err := vm.Run(JavaScript)
			if err != nil {
				log.Println("解析图片地址失败:", err)
			}
			imgStr, err := value.ToString()
			if err != nil {
				log.Println("图片地址解析出错:", err)
			}
			imgArr := strings.Split(imgStr, ",")
			catalog.ImgArr = imgArr[1:]
			if len(catalog.ImgArr) == 0 {
				log.Println("图片获取失败..............")
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
	})

	c.Visit(catalog.Url)
}

// 创建文件夹并存储图片
func CreateFileGetImg(catalog *Catalog, index int) {
	// 创建文件目录
	if len(catalog.ImgArr) == 0 {
		log.Println(catalog.Title, "图片目录为空")
		return
	}
	//dir := Title + "/" + +"--" + catalog.Title
	file := Title + "/" + strconv.Itoa(index) + "-"
	//err := os.MkdirAll(dir, os.ModePerm)
	//if err != nil {
	//	log.Println("创建文件夹失败")
	//}

	// 下载图片
	for k, v := range catalog.ImgArr {
	GetImage:
		resp, err := http.Get(ImgHeader + v)
		if err != nil {
			log.Println(err)
			goto GetImage // 获取图片出错重新获取
		}
		body, _ := ioutil.ReadAll(resp.Body)
		out, _ := os.Create(file + strconv.Itoa(k+1) + "--" + catalog.Title + ".jpg")
		io.Copy(out, bytes.NewReader(body))
		resp.Body.Close()
	}
}

func coverGBKToUTF8(src string) string {
	return mahonia.NewDecoder("gbk").ConvertString(src)
}

type pool struct {
	queue chan int
	wg    *sync.WaitGroup
}

func New(size int) *pool {
	if size <= 0 {
		size = 1
	}
	return &pool{
		queue: make(chan int, size),
		wg:    &sync.WaitGroup{},
	}
}

func (p *pool) Add(delta int) {
	for i := 0; i < delta; i++ {
		p.queue <- 1
	}
	for i := 0; i > delta; i-- {
		<-p.queue
	}
	p.wg.Add(delta)
}

func (p *pool) Done() {
	<-p.queue
	p.wg.Done()
}

func (p *pool) Wait() {
	p.wg.Wait()
}
