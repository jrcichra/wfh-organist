package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jrcichra/wfh-organist/internal/types"
	"github.com/tidwall/gjson"
)

// https://tutorialedge.net/golang/go-websocket-tutorial/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (s *Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	key := ws.RemoteAddr().String()
	log.Println("Websocket Client Connected!", key)
	s.websocketsMutex.Lock()
	s.websockets[key] = ws
	s.websocketsMutex.Unlock()
	s.reader(ws)
}

func (s *Server) reader(conn *websocket.Conn) {
	for {
		// read in a message
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		json := string(p)

		// extract fields
		typ := gjson.Get(json, "type")

		switch typ.String() {
		case "noteOn":
			s.notesChan <- types.NoteOn{
				Time:     time.Now(),
				Key:      uint8(gjson.Get(json, "key").Uint()),
				Velocity: uint8(gjson.Get(json, "velocity").Uint()),
				Channel:  uint8(gjson.Get(json, "channel").Uint()),
			}
		case "noteOff":
			s.notesChan <- types.NoteOff{
				Time:    time.Now(),
				Key:     uint8(gjson.Get(json, "key").Uint()),
				Channel: uint8(gjson.Get(json, "channel").Uint()),
			}
		default:
			log.Println("Unknown message type:", typ.String())
		}
	}
}
