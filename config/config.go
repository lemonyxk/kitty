package config

import (
	"io/ioutil"
	"path/filepath"
)

type Config struct {
	bytes []byte
	dir   string
	file  string
}

func (c *Config) SetConfigFile(configFile string) error {

	absPath, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}

	c.dir = filepath.Dir(absPath)
	c.file = absPath
	c.bytes = bytes

	return nil
}

func (c *Config) Bytes() []byte {
	return c.bytes
}

func (c *Config) Dir() string {
	return c.dir
}

func (c *Config) File() string {
	return c.file
}
