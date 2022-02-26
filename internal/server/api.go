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
func handleAPI(notesChan chan interface{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("API request:", r.URL.Path)
		switch r.URL.Path {
		// case "/api/midi/noteon":
		// 	apiHandleNoteOn(w, r)
		// case "/api/midi/noteoff":
		// 	apiHandleNoteOff(w, r)
		// case "/api/midi/programchange":
		// 	apiHandleProgramChange(w, r)
		// case "/api/midi/aftertouch":
		// 	apiHandleAfterTouch(w, r)
		// case "/api/midi/controlchange":
		// 	apiHandleControlChange(w, r)
		// case "/api/midi/pitchbend":
		// 	apiHandlePitchBend(w, r)
		// case "/api/midi/polyaftertouch":
		// 	apiHandlePolyAfterTouch(w, r)
		case "/api/midi/raw":
			apiHandleRaw(w, r, notesChan)
		case "/api/midi/files":
			apiHandleStat(w, r)
		case "/api/midi/file/play":
			apiHandlePlay(w, r, notesChan)
		case "/api/midi/file/stop":
			apiHandleStop(w, r, notesChan)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

var stopPlayingChan = make(chan struct{})

func apiHandleStat(w http.ResponseWriter, r *http.Request) {
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

func apiHandleStop(w http.ResponseWriter, r *http.Request, notesChan chan interface{}) {
	select {
	case stopPlayingChan <- struct{}{}:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func apiHandlePlay(w http.ResponseWriter, r *http.Request, notesChan chan interface{}) {
	// make sure it's a post
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
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
	go player.PlayMidiFile(notesChan, "midi/"+filename, stopPlayingChan)
	// send a success message
	w.WriteHeader(http.StatusOK)
}

func apiHandleRaw(w http.ResponseWriter, r *http.Request, notesChan chan interface{}) {
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
			notesChan <- types.Raw{
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
