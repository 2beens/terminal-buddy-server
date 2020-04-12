package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type TBConfig struct {
	Env string `yaml:"env"`

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
