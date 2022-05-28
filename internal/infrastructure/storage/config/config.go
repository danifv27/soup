package config

import (
	"io/ioutil"

	"github.com/danifv27/soup/internal/domain/soup"
	"gopkg.in/yaml.v2"
)

type SoupConfig struct {
	soupinfo soup.SoupInfo
}

//NewSoupConfig implements config interface
func NewSoupConfig(cloneLocation string) SoupConfig {

	info := new(soup.SoupInfo)
	return SoupConfig{
		soupinfo: *info,
	}
}

// GetSoupInfo read and decrypts .soup.yaml file
func (s SoupConfig) GetSoupInfo(root string) soup.SoupInfo {

	yamlFile, err := ioutil.ReadFile(root + "/.soup.yaml")
	if err != nil {
		return soup.SoupInfo{}
	}

	err = yaml.Unmarshal(yamlFile, &s.soupinfo)
	if err != nil {
		return soup.SoupInfo{}
	}

	s.soupinfo.Root = root

	return s.soupinfo
}
