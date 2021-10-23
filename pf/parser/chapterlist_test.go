package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestParseChapterList(t *testing.T) {
	file, err := os.Open("./index.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	data := ParseChapterList(b)
	for _, item := range data.Items {
		fmt.Printf("Got item %v\n", item)
	}
}
