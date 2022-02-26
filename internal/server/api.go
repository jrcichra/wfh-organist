package server

import (
	"bufio"
	"encoding/hex"
	"log"
	"net/http"
	"time"

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
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
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
