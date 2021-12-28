package main

import (
	"log"

	"gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/rtmididrv"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func cont(err error) {
	if err != nil {
		log.Println(err)
	}
}

func printPort(port midi.Port) {
	log.Printf("[%v] %s\n", port.Number(), port.String())
}

func printOutPorts(ports []midi.Out) {
	log.Printf("MIDI OUT Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	log.Printf("\n\n")
}

func printInPorts(ports []midi.In) {
	log.Printf("MIDI IN Ports\n")
	for _, port := range ports {
		printPort(port)
	}
	log.Printf("\n\n")
}

func getLists() {
	drv, err := driver.New()
	must(err)

	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	printInPorts(ins)
	printOutPorts(outs)
}
