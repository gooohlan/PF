package parser

import (
	"regexp"
	"strconv"

	"PF/common"
	"PF/engine"
)

var reChapterList = regexp.MustCompile(`<li><a href="/(manhua/[0-9]+/[0-9]+.html)" title="[^>]+">([^<]+)</a></li>`)
var reTitle = regexp.MustCompile(`<div class="titleInfo"><h1>([^<]+)</h1><span>连载</span></div>`)

func ParseChapterList(contents []byte) engine.ParseResult {
	matches := reChapterList.FindAllSubmatch(contents, -1)
	result := engine.ParseResult{}
	for i, m := range matches {
		result.Items = append(result.Items, "Chapter List "+string(m[2]))
		result.Requests = append(result.Requests, engine.Request{
			Url:    common.UrlJoin(engine.PFUrl, string(m[1])),
			Parser: NewComicListParser(strconv.Itoa(len(matches)-i) + "." + string(m[2])),
		})
	}
	title := reTitle.FindAllSubmatch(contents, -1)
	if len(title) > 0 {
		engine.ComicName = string(title[0][1])
		common.CreateDirectory(engine.ComicName)
	}
	return result
}
