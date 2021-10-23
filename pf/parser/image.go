package parser

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"

	"PF/engine"
)

func ParseImage(content []byte, name string) engine.ParseResult {
	log.Println("create image name:", name)
	result := engine.ParseResult{}
	out, err := os.Create(filepath.Join(engine.ComicName, name) + ".jpg")
	if err != nil {
		log.Printf("create image err: %v", err)
		return result
	}
	_, err = io.Copy(out, bytes.NewReader(content))
	if err != nil {
		log.Printf("write image error:%v", err)
		return result
	}
	return result
}


type ImageParser struct {
	name string
}

func NewImageParser(name string) *ImageParser  {
	return &ImageParser{name: name}
}

func (i ImageParser) Parse(contents []byte) engine.ParseResult {
	return ParseImage(contents, i.name)
}