package server

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/player"
	"github.com/jrcichra/wfh-organist/internal/state"
	"github.com/jrcichra/wfh-organist/internal/types"
)

// send message from the api to the midi server

// handle all the API endpoints
func (s *Server) handleAPI() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("API request:", r.URL.Path)
		switch r.URL.Path {
		case "/api/midi/pushstop":
			s.apiHandlePushStop(w, r)
		case "/api/midi/files":
			s.apiHandleStat(w, r)
		case "/api/midi/file/play":
			s.apiHandlePlay(w, r)
		case "/api/midi/file/stop":
			s.apiHandleStopButton(w, r)
		case "/api/midi/stops":
			s.apiStops(w, r)
		case "/api/midi/panic":
			// s.apiHandlePanic(w, r)
		case "/api/midi/piston":
			s.apiHandlePiston(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

var stopPlayingChan = make(chan bool)

func (s *Server) apiHandlePiston(w http.ResponseWriter, r *http.Request) {
	// make sure it's a post
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get the piston from the body
	scanner := bufio.NewScanner(r.Body)
	if !scanner.Scan() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pistonStr := scanner.Text()

	piston, err := strconv.Atoi(pistonStr)
	if err != nil {
		common.Cont(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// piston 0 = the cancel button

	stops := s.state.GetPiston(piston)

	// send a program change to channels 0-2 (aka. human 1-3)
	// the value of the last byte is the piston number
	for i := 0; i < 3; i++ {
		s.notesChan <- types.ProgramChange{
			Time:    time.Now(),
			Channel: uint8(i),
			Program: uint8(piston),
		}
	}

	// tell notes chan what stops to press
	for _, stop := range stops {
		// get the current state and compare to the desired state
		pressed, err := s.state.GetStopAPI(stop)
		common.Cont(err)
		if pressed && !stop.Pressed {
			s.state.SetStopAPI(stop, false)
		} else if !pressed && stop.Pressed {
			s.state.SetStopAPI(stop, true)
		}
	}

	// send json of new stops
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stops)
}

func (s *Server) apiStops(w http.ResponseWriter, r *http.Request) {
	// make sure it's a get
	switch r.Method {
	case "GET":
		// convert stops to json and then send
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.state.GetStopsForAPI())
	case "POST":
		w.WriteHeader(http.StatusOK)
		// parse the json response (which is an array of APIStops)
		var stops []state.APIStop
		err := json.NewDecoder(r.Body).Decode(&stops)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// get the query string
		query := r.URL.Query()
		// get the piston
		pistonStr := query.Get("piston")

		// update the state of the piston if there is a piston
		if pistonStr != "" {
			piston, err := strconv.Atoi(pistonStr)
			if err != nil {
				common.Cont(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.state.SetPiston(piston, stops)
		} else {
			// otherwise just update the current state
			for _, stop := range stops {
				pressed, err := s.state.GetStopAPI(stop)
				common.Cont(err)
				if pressed && !stop.Pressed {
					s.state.SetStopAPI(stop, false)
				} else if !pressed && stop.Pressed {
					s.state.SetStopAPI(stop, true)
				}
			}
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) apiHandleStat(w http.ResponseWriter, r *http.Request) {
	// get a list of midi files in the midi directory
	log.Println("Getting list of midi files")
	matches, err := filepath.Glob("midi/*")
	if err != nil {
		log.Println(err)
	}

	// only get the basenames
	files := make([]string, len(matches))
	for i, match := range matches {
		files[i] = filepath.Base(match)
	}

	// send the list of files in JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *Server) apiHandleStopButton(w http.ResponseWriter, r *http.Request) {
	select {
	case stopPlayingChan <- true:
		time.Sleep(time.Millisecond * 500)
		// send all notes off
		s.notesChan <- types.Raw{
			Time: time.Now(),
			Data: []byte{0xB0, 0x7B, 0x00},
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Server) apiHandlePlay(w http.ResponseWriter, r *http.Request) {
	// make sure it's a post
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// stop anything in progress
	select {
	case stopPlayingChan <- true:
	default:
	}

	// get the filename from the body
	scanner := bufio.NewScanner(r.Body)
	if !scanner.Scan() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	filename := scanner.Text()
	// start a player that opens the filename specified
	go player.PlayMidiFile(s.notesChan, "midi/"+filename, stopPlayingChan, true)
	// send a success message
	w.WriteHeader(http.StatusOK)
}

func (s *Server) apiHandlePushStop(w http.ResponseWriter, r *http.Request) {
	// make sure it's a post
	if r.Method != "POST" {
		log.Println("Not a POST request")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// get the id for the stop from the body
	scanner := bufio.NewScanner(r.Body)
	if !scanner.Scan() {
		log.Println("No id specified")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := scanner.Text()

	pressed, err := s.state.GetStopFromID(id)
	if err != nil {
		common.Cont(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// toggle the state of the press
	err = s.state.SetStopFromID(id, !pressed)
	if err != nil {
		common.Cont(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// send a success message
	w.WriteHeader(http.StatusOK)

}
