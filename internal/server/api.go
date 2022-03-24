package server

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jrcichra/wfh-organist/internal/player"
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
			s.apiHandleStop(w, r)
		case "/api/midi/stops":
			s.apiGetStops(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

var stopPlayingChan = make(chan bool)

func (s *Server) apiGetStops(w http.ResponseWriter, r *http.Request) {
	// make sure it's a get
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// convert stops to json and then send
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.state.GetStopsForAPI())

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

func (s *Server) apiHandleStop(w http.ResponseWriter, r *http.Request) {
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
	scanner.Split(bufio.ScanWords)
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
	scanner.Split(bufio.ScanWords)
	if !scanner.Scan() {
		log.Println("No id specified")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := scanner.Text()

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the stop string from the state
	stop, pressed := s.state.GetStop(intID)
	if stop == "" {
		log.Println("Stop not found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// split the stop by whitespace
	byteStrSets := strings.Split(stop, " ")
	var bytes []byte
	for _, byteStr := range byteStrSets {
		bite, err := hex.DecodeString(byteStr)
		if err != nil {
			log.Println("Invalid hex string")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		bytes = append(bytes, bite...)
	}

	// add final byte opposite of pressed status
	if pressed {
		bytes = append(bytes, 0x00)
	} else {
		bytes = append(bytes, 0x7f)
	}

	// send in chunks of 3
	for i := 0; i < len(bytes); i += 3 {
		// send the stop to the notes channel
		s.notesChan <- types.Raw{
			Time: time.Now(),
			Data: bytes[i : i+3],
		}
	}

	// toggle the state of the press
	s.state.ToggleStop(intID)

	// send a success message
	w.WriteHeader(http.StatusOK)

}
