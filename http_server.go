package lemo

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

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

func (h *HttpServer) staticHandler(filePath string) error {

	log.Println(exec.LookPath(os.Args[0]))
	return nil

}
