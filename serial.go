package main

import (
	"bufio"
	"fmt"

	"github.com/tarm/serial"
)

/*

// channels 0 1 and 2 need to change (my numbers) human 1 2 3
Control b0 b1 and b2 all at the same time
data is 07
starts at 97 decimal for the value. maybe htat's where the pedal was left
looks like 42 decimal is the lowest value. Seeing numbers separated by about 4.
127 is the highest value.

*/

// read the serial port and

func readSerial() {
	c := &serial.Config{Name: "/dev/ttyACM0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		cont(err)
	} else {
		scanner := bufio.NewScanner(s)
		for scanner.Scan() {
			fmt.Println("Expression:", scanner.Text())
		}
	}
}
