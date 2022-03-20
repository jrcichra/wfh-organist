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

type Stops struct {
	Swell   Group `yaml:"swell" json:"swell"`
	Great   Group `yaml:"great" json:"great"`
	Pedal   Group `yaml:"pedal" json:"pedal"`
	General Group `yaml:"general" json:"general"`
}

func ReadFile(filename string) *Stops {
	data, err := ioutil.ReadFile(filename)
	common.Must(err)
	stops := &Stops{}
	err = yaml.Unmarshal(data, stops)
	common.Must(err)
	return stops
}
