package main

import (
	"io/ioutil"
	"log"
	"net/http"
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
	// read the body into a string
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// parse the body
	raw, err := hexToRawStruct(string(body))
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// send the message
	notesChan <- raw
	// send a success message
	w.WriteHeader(http.StatusOK)

}
