package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type LogOutput int

const (
	StdoutLogOutput LogOutput = iota
	FileLogOutput
)

type EnvConfig struct {
	Port int

	Log struct {
		Out   string
		File  string
		Level string
	}

	DB struct {
		User string
		Name string
	}
}

type TBConfig struct {
	Env        string    `yaml:"env"`
	Production EnvConfig `yaml:"production"`
	Dev        EnvConfig `yaml:"dev"`
}

func NewTbConfig(configData []byte) (*TBConfig, error) {
	tbConfig := &TBConfig{}
	err := yaml.Unmarshal(configData, tbConfig)
	if err != nil {
		return nil, err
	}
	return tbConfig, nil
}

func ReadYamlConfig(path string) ([]byte, error) {
	yamlConfFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := yamlConfFile.Close()
		if err != nil {
			panic(fmt.Errorf("read yaml config - close config error: %s", err))
		}
	}()

	return ioutil.ReadAll(yamlConfFile)
}

func (c *TBConfig) Port() int {
	if c.Env == "prod" {
		return c.Production.Port
	}
	return c.Dev.Port
}

func (c *TBConfig) LogOutput() LogOutput {
	if c.Env == "prod" {
		if c.Production.Log.Out == "file" {
			return FileLogOutput
		}
	}
	if c.Env == "dev" {
		if c.Dev.Log.Out == "file" {
			return FileLogOutput
		}
	}
	return StdoutLogOutput
}

func (c *TBConfig) LogFilePath() string {
	if c.Env == "prod" {
		return c.Production.Log.File
	}
	return c.Dev.Log.File
}

func (c *TBConfig) LogLevel() string {
	if c.Env == "prod" {
		return c.Production.Log.Level
	}
	return c.Dev.Log.Level
}

func (c *TBConfig) DbName() string {
	if c.Env == "prod" {
		return c.Production.DB.Name
	}
	return c.Dev.DB.Name
}

func (c *TBConfig) DbUser() string {
	if c.Env == "prod" {
		return c.Production.DB.User
	}
	return c.Dev.DB.User
}
