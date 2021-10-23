package common

import (
	"net/url"
	"os"
	"path"
)

func UrlJoin(faceUrl, elem string) string {
	u, _ := url.Parse(faceUrl)
	u.Path = path.Join(u.Path, elem)
	return u.String()
}

func CreateDirectory(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	err = os.Mkdir(path, os.ModePerm)
	return err
}