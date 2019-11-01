package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/tidwall/gjson"

	"github.com/Lemo-yxk/lemo"
)

type Config struct {
	result gjson.Result
	dir    string
	file   string
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
	c.result = gjson.ParseBytes(bytes)

	return nil
}

// GetByID 获取
func (c *Config) Result() gjson.Result {
	return c.result
}

func (c *Config) Path(path string) gjson.Result {
	return c.result.Get(path)
}

func (c *Config) JsonString() string {
	return c.result.Raw
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) File() string {
	return c.file
}

func StringArray(arrayResult []gjson.Result) {
	var result []string
	for _, value := range arrayResult {
		result = append(result, value.String())
	}
}
