package stops

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

func ReadFile(filename string) *Config {
	data, err := ioutil.ReadFile(filename)
	common.Must(err)
	stops := &Config{}
	err = yaml.Unmarshal(data, stops)
	common.Must(err)
	return stops
}
