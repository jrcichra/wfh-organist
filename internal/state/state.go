package state

import (
	"encoding/hex"
	"errors"
	"fmt"
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
	// log.Println("kvPut", key, value)
	return s.db.Put([]byte(key), []byte(value))
}

func (s *State) kvGet(key string) (string, error) {
	// log.Println("kvGet", key)
	res, err := s.db.Get([]byte(key))
	if err != nil && strings.Contains(err.Error(), "key not found") {
		return "false", nil
	}
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func (s *State) SetStopAPI(providedStop APIStop, value bool) {
	id := s.GetStopAPIID(providedStop)
	err := s.kvPut(id, strconv.FormatBool(value))
	common.Cont(err)

	code, err := s.GetStopCode(providedStop)
	common.Cont(err)

	_, err = s.codeToNotesChan(id, code, value)
	common.Cont(err)

}

func (s *State) convertPressed(pressed string, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	pressedBool, err := strconv.ParseBool(pressed)
	if err != nil {
		return false, err
	}
	return pressedBool, nil
}

func (s *State) GetStopAPI(providedStop APIStop) (bool, error) {
	return s.convertPressed(s.kvGet(s.GetStopAPIID(providedStop)))
}

func (s *State) GetStop(providedStop config.Stop) (bool, error) {
	return s.convertPressed(s.kvGet(s.GetStopID(providedStop)))
}

func (s *State) GetStopCode(providedStop APIStop) (string, error) {
	id := s.GetStopAPIID(providedStop)
	for _, stop := range s.config.Stops {
		if s.GetStopID(stop) == id {
			return stop.Code, nil
		}
	}
	return "", errors.New("stop " + id + " not found")
}

func (s *State) GetStopCodeFromID(id string) (string, error) {
	for _, stop := range s.config.Stops {
		if s.GetStopID(stop) == id {
			return stop.Code, nil
		}
	}
	return "", errors.New("stop " + id + " not found")
}

func (s *State) GetStopsForAPI() []APIStop {

	stops := s.config.Stops
	apiStops := make([]APIStop, len(stops))

	for i, stop := range stops {
		pressed, err := s.GetStop(stop)
		common.Cont(err)
		apiStops[i] = APIStop{Name: stop.Name, Group: stop.Group, Pressed: pressed}
	}

	return apiStops
}

func (s *State) SetPiston(piston int, stops []APIStop) {
	for _, stop := range stops {
		id := s.GetPistonAPIID(piston, stop)
		err := s.kvPut(id, strconv.FormatBool(stop.Pressed))
		common.Cont(err)
	}
}

func (s *State) GetStopAPIID(stop APIStop) string {
	return fmt.Sprintf("stop/%s/%s", stop.Group, stop.Name)
}

func (s *State) GetStopID(stop config.Stop) string {
	return fmt.Sprintf("stop/%s/%s", stop.Group, stop.Name)
}

func (s *State) GetPistonAPIID(piston int, stop APIStop) string {
	return fmt.Sprintf("piston/%d/%s", piston, s.GetStopAPIID(stop))
}
func (s *State) GetPistonID(piston int, stop config.Stop) string {
	return fmt.Sprintf("piston/%d/%s", piston, s.GetStopID(stop))
}

func (s *State) GetPistonStop(piston int, stop config.Stop) (bool, error) {
	id := s.GetPistonID(piston, stop)
	return s.convertPressed(s.kvGet(id))
}

func (s *State) GetPiston(piston int) []APIStop {
	stops := s.config.Stops
	apiStops := make([]APIStop, len(stops))

	for i, stop := range stops {
		pressed, err := s.GetPistonStop(piston, stop)
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

func (s *State) GetStopPressedFromID(id string) (bool, error) {
	return s.convertPressed(s.kvGet(id))
}

func (s *State) SetStopPressedFromID(id string, value bool) error {
	err := s.kvPut(id, strconv.FormatBool(value))
	common.Cont(err)

	code, err := s.GetStopCodeFromID(id)
	common.Cont(err)

	_, err = s.codeToNotesChan(id, code, value)
	common.Cont(err)
	return err
}
