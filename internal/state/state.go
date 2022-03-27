package state

import (
	"encoding/hex"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"git.mills.io/prologic/bitcask"
	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/parser/config"
	"github.com/jrcichra/wfh-organist/internal/types"
)

type APIStop struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Pressed bool   `json:"pressed"`
}

// State couples the state of the GUIs, the config, and the notes channel
type State struct {
	db        *bitcask.Bitcask
	profile   string
	config    *config.Config
	notesChan chan interface{}
}

func (s *State) Open(profile string, notesChan chan interface{}) {
	var err error
	s.profile = profile
	s.notesChan = notesChan
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
	log.Println("Setting", id, "pressed to", value)
	err := s.kvPut("pressed/"+id, strconv.FormatBool(value))
	common.Cont(err)

	code, err := s.GetStopCode(id)
	common.Cont(err)

	_, err = s.codeToNotesChan(id, code, value)
	common.Cont(err)

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

func (s *State) SetPiston(piston string, stops []APIStop) {
	for _, stop := range stops {
		s.SetPressed("piston/"+piston+"/"+stop.Group+"/"+stop.Name, stop.Pressed)
	}
}

func (s *State) GetPiston(piston string) []APIStop {
	stops := s.config.Stops
	apiStops := make([]APIStop, len(stops))

	for i, stop := range stops {
		pressed, err := s.GetPressed("piston/" + piston + "/" + stop.Group + "/" + stop.Name)
		common.Cont(err)
		apiStops[i] = APIStop{Name: stop.Name, Group: stop.Group, Pressed: pressed}
	}

	return apiStops
}

func (s *State) codeToNotesChan(id string, code string, pressed bool) (bool, error) {
	// split the stop code by whitespace
	byteStrSets := strings.Split(code, " ")
	var bytes []byte
	for _, byteStr := range byteStrSets {
		bite, err := hex.DecodeString(byteStr)
		if err != nil {
			return false, err
		}
		bytes = append(bytes, bite...)
	}

	if pressed {
		bytes = append(bytes, 0x7f)
	} else {
		bytes = append(bytes, 0x00)
	}

	// send in chunks of 3
	for i := 0; i < len(bytes); i += 3 {
		// send the stop to the notes channel
		s.notesChan <- types.Raw{
			Time: time.Now(),
			Data: bytes[i : i+3],
		}
	}

	return pressed, nil
}
