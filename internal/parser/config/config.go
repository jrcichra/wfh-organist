package config

import (
	"io/ioutil"

	"github.com/jrcichra/wfh-organist/internal/common"
	"gopkg.in/yaml.v2"
)

type Stop struct {
	Name  string `yaml:"name" json:"name"`
	Code  string `yaml:"code" json:"code"`
	Group string `yaml:"group" json:"group"`
}

type Group []Stop

type Config struct {
	Stops []Stop `yaml:"stops" json:"stops"`
}

func (c *Config) Read(profile string) {

	type File struct {
		Stops []map[string]Group `yaml:"stops" json:"stops"`
	}

	var f File

	data, err := ioutil.ReadFile(profile + "/stops.yaml")
	common.Must(err)
	err = yaml.Unmarshal(data, &f)
	common.Must(err)

	// restructure the data
	for _, top := range f.Stops {
		for groupName, group := range top {
			for _, stop := range group {
				c.Stops = append(c.Stops, Stop{Name: stop.Name, Code: stop.Code, Group: groupName})
			}
		}
	}
}

func (c *Config) ReadString(data string) {
	err := yaml.Unmarshal([]byte(data), c)
	common.Cont(err)
}
