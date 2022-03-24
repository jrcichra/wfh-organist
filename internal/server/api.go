package server

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
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
		case "/api/midi/raw":
			s.apiHandleRaw(w, r)
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
	json.NewEncoder(w).Encode(s.stops) // expects *config.Config

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

func (s *Server) apiHandleRaw(w http.ResponseWriter, r *http.Request) {
	// make sure it's a post
	if r.Method != "POST" {
		log.Println("Not a POST request")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	scanner := bufio.NewScanner(r.Body)

	scanner.Split(bufio.ScanWords)
	// keep count, send 3 at a time
	count := 0
	// hold bytes
	bytes := make([]byte, 0)
	for scanner.Scan() {
		//convert token string to hex code
		text := scanner.Text()
		// each token must be size 2
		if len(text) != 2 {
			panic("Token must be size 2")
		}
		hexToken, err := hex.DecodeString(text)
		if err != nil {
			log.Println(err)
			break
		}
		// append to bytes
		bytes = append(bytes, hexToken...)
		if count >= 2 {
			//send hex code to channel
			s.notesChan <- types.Raw{
				Time: time.Now(),
				Data: bytes,
			}
			count = 0
			// clear bytes
			bytes = make([]byte, 0)
		} else {
			count++
		}
	}
	// send a success message
	w.WriteHeader(http.StatusOK)

}
