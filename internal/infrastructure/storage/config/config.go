package config

import (
	"io/ioutil"

	"github.com/danifv27/soup/internal/domain/soup"
	"gopkg.in/yaml.v2"
)

type SoupRepo struct {
	soupinfo soup.SoupInfo
}

//NewSoupRepo Constructor
func NewSoupRepo(cloneLocation string) SoupRepo {

	info := new(soup.SoupInfo)
	return SoupRepo{
		soupinfo: *info,
	}
}

func (s SoupRepo) GetSoupInfo(root string) soup.SoupInfo {

	yamlFile, err := ioutil.ReadFile(root + "/.soup.yaml")
	if err != nil {
		return soup.SoupInfo{}
	}

	err = yaml.Unmarshal(yamlFile, &s.soupinfo)
	if err != nil {
		return soup.SoupInfo{}
	}

	return s.soupinfo
}
