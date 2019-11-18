package utils

import (
	"github.com/json-iterator/go"
	"io/ioutil"
	"path/filepath"

	"github.com/Lemo-yxk/lemo"
)

type Config struct {
	bytes []byte
	dir   string
	file  string
	any   jsoniter.Any
}

func (c *Config) SetConfigFile(configFile string) func() *lemo.Error {

	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return lemo.NewError(err)
	}
	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return lemo.NewError(err)
	}

	c.dir = filepath.Dir(absPath)
	c.file = absPath
	c.bytes = bytes
	c.any = jsoniter.Get(c.bytes)

	return nil
}

// GetByID 获取
func (c *Config) Any() jsoniter.Any {
	return c.any
}

func (c *Config) Bytes() []byte {
	return c.bytes
}

func (c *Config) Path(path string) jsoniter.Any {
	return c.any.Get(path)
}

func (c *Config) JsonString() string {
	return c.any.ToString()
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) File() string {
	return c.file
}

func (c *Config) ArrayString(path string) []string {
	var result []string
	var val = c.any.Get(path)
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.ToString())
	}
	return result
}
