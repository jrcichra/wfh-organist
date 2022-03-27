package state

import (
	"errors"
	"strconv"
	"strings"

	"git.mills.io/prologic/bitcask"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/parser/config"
)

type APIStop struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Pressed bool   `json:"pressed"`
}

type State struct {
	db      *bitcask.Bitcask
	profile string
	config  *config.Config
}

func (s *State) Open(profile string) {
	var err error
	s.profile = profile
	s.db, err = bitcask.Open(s.profile + "/state")
	if err != nil {
		common.Must(err)
	}
	s.config = &config.Config{}
	s.config.Read(s.profile)
}

func (s *State) kvPut(key string, value string) error {
	return s.db.Put([]byte(key), []byte(value))
}

func (s *State) kvGet(key string) (string, error) {
	res, err := s.db.Get([]byte(key))
	if err != nil && strings.Contains(err.Error(), "key not found") {
		return "false", nil
	}
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (s *State) SetPressed(id string, value bool) {
	err := s.kvPut("pressed/"+id, strconv.FormatBool(value))
	common.Must(err)
}

func (s *State) GetPressed(id string) (bool, error) {
	pressed, err := s.kvGet("pressed/" + id)
	if err != nil {
		return false, err
	}
	pressedBool, err := strconv.ParseBool(pressed)
	if err != nil {
		return false, err
	}
	return pressedBool, nil
}

func (s *State) GetStopCode(id string) (string, error) {
	for _, stop := range s.config.Stops {
		if stop.Group+"/"+stop.Name == id {
			return stop.Code, nil
		}
	}
	return "", errors.New("stop " + id + " not found")
}

func (s *State) GetStopsForAPI() []APIStop {

	stops := s.config.Stops
	apiStops := make([]APIStop, len(stops))

	for i, stop := range stops {
		pressed, err := s.GetPressed(stop.Group + "/" + stop.Name)
		common.Cont(err)
		apiStops[i] = APIStop{Name: stop.Name, Group: stop.Group, Pressed: pressed}
	}

	return apiStops
}
