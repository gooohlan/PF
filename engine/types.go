package engine

const (
	PFUrl  = "http://pufei.org/"
	ImgUrl = "http://res.img.jituoli.com/"
)

var ComicName string

type ParserFunc func([]byte) ParseResult

type Parser interface {
	Parse(contents []byte) ParseResult
}

type ParseResult struct {
	Requests []Request
	Items    []interface{}
}

const (
	Html = iota
	Image
)

type Request struct {
	Url    string
	Parser Parser
	Type   int
}

type FuncParser struct {
	parser ParserFunc
}

func NewFuncParser(parser ParserFunc) *FuncParser {
	return &FuncParser{parser: parser}
}

func (f *FuncParser) Parse(
	contents []byte) ParseResult {
	return f.parser(contents)
}

type NilParse struct{}

func NewNilParse() *NilParse {
	return &NilParse{}
}

func (n NilParse) Parse(_ []byte) ParseResult {
	return ParseResult{}
}
