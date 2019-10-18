package lemo

import (
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lemo-yxk/tire"
)

type HttpServer struct {
	IgnoreCase bool
	Router     *tire.Tire
	OnError    ErrorFunction

	group      *httpServerGroup
	route      *httpServerRoute
	prefixPath string
	staticPath string
}

func (h *HttpServer) Ready() {

}

func (h *HttpServer) SetStaticPath(prefixPath string, staticPath string) {

	if prefixPath == "" {
		panic("prefixPath can not be empty")
	}

	if staticPath == "" {
		panic("staticPath can not be empty")
	}

	absStaticPath, err := filepath.Abs(staticPath)
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(absStaticPath)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic("staticPath is not a dir")
	}

	h.prefixPath = prefixPath
	h.staticPath = absStaticPath
}

func (h *HttpServer) staticHandler(filePath string) ([]byte, string, func() *Error) {

	if !strings.HasPrefix(filePath, h.prefixPath) {
		return nil, "", NewError("not match")
	}

	var absFilePath = h.staticPath + filePath[len(h.prefixPath):]

	var info, err = os.Stat(absFilePath)
	if err != nil {
		return nil, "", NewError(err)
	}

	if info.IsDir() {
		absFilePath = filepath.Join(absFilePath, "index.html")
		if _, err := os.Stat(absFilePath); err != nil {
			return nil, "", NewError("staticPath is not a file")
		}
	}

	// has found
	var contentType = mime.TypeByExtension(filepath.Ext(absFilePath))

	f, err := os.OpenFile(absFilePath, os.O_RDONLY, 0666)
	if err != nil {
		return nil, contentType, NewError(err)
	}

	bts, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, contentType, NewError(err)
	}

	return bts, contentType, nil

}
