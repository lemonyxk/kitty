package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/json-iterator/go"

	"github.com/Lemo-yxk/lemo/exception"
)

type Config struct {
	bytes []byte
	dir   string
	file  string
	any   jsoniter.Any
}

func (c *Config) SetConfigFile(configFile string) exception.ErrorFunc {

	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return exception.New(err)
	}
	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return exception.New(err)
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

func (c *Config) Path(path ...interface{}) jsoniter.Any {
	return c.any.Get(path...)
}

func (c *Config) String() string {
	return c.any.ToString()
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) File() string {
	return c.file
}

func (c *Config) ArrayString(path ...interface{}) []string {
	var result []string
	var val = c.any.Get(path...)
	for i := 0; i < val.Size(); i++ {
		result = append(result, val.Get(i).ToString())
	}
	return result
}
