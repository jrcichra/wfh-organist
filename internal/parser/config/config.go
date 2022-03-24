package config

import (
	"io/ioutil"

	"github.com/jrcichra/wfh-organist/internal/common"
	"gopkg.in/yaml.v2"
)

type Stop struct {
	Name string `yaml:"name" json:"name"`
	Code string `yaml:"code" json:"code"`
}

type Group []Stop

type Config struct {
	Stops []map[string]Group `yaml:"stops" json:"stops"`
}

func (c *Config) Read(profile string) {
	data, err := ioutil.ReadFile(profile + "/stops.yaml")
	common.Must(err)
	err = yaml.Unmarshal(data, c)
	common.Must(err)
}
