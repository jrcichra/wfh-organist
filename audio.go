package main

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"syscall"

	"github.com/gordonklaus/portaudio"
)

const sampleRate = 16000
const bufferSize = 64

/*
numbers that are known to work
sampleRate        bufferSizes
16000             64
44100             2,8
*/

func audioServer(port int) {

	// wait for someone to connect to the server
	l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	must(err)
	log.Println("Listening for audio connection on port", port)
	for {
		conn, err := l.Accept()
		log.Println("audio client connected")
		go func() {
			must(err)
			portaudio.Initialize()
			defer portaudio.Terminate()
			in := make([]byte, bufferSize)
			stream, err := portaudio.OpenDefaultStream(2, 0, sampleRate, len(in), in)
			must(err)
			defer stream.Close()
			must(stream.Start())
			for {
				cont(stream.Read())
				err := binary.Write(conn, binary.LittleEndian, in)
				if errors.Is(err, syscall.EPIPE) {
					break
				}
				cont(err)
			}
			must(stream.Stop())
			log.Println("audio client disconnected")
		}()
	}
}

func audioClient(server string, port int) {

	conn, err := net.Dial("tcp", server+":"+strconv.Itoa(port))
	must(err)

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]byte, bufferSize)
	stream, err := portaudio.OpenDefaultStream(0, 2, sampleRate, len(in), in)
	must(err)
	defer stream.Close()
	must(stream.Start())
	for {
		err := binary.Read(conn, binary.LittleEndian, in)
		if err == io.EOF {
			break
		}
		cont(err)
		cont(stream.Write())
	}
	must(stream.Stop())
}
