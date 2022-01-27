package main

import (
	"bufio"
	"log"
	"strconv"
	"time"

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

// read the serial port assuming it's the expression pedal
func readSerial(notesChan chan interface{}) {
	c := &serial.Config{Name: "/dev/ttyACM0", Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		cont(err)
	} else {
		scanner := bufio.NewScanner(s)
		pPercentage := 0
		for scanner.Scan() {
			expression, err := strconv.Atoi(scanner.Text())
			if err != nil {
				cont(err)
			}
			percentage := expressionPercentage(expression)
			if percentage != pPercentage {
				log.Println("percentage before math", percentage)
				value := float64(percentage) / 100.0 * 127.0
				log.Println("expression value", value)
				notesChan <- Raw{Time: time.Now(), Data: []byte{0xB0, 0x07, uint8(value)}}
				notesChan <- Raw{Time: time.Now(), Data: []byte{0xB1, 0x07, uint8(value)}}
				notesChan <- Raw{Time: time.Now(), Data: []byte{0xB2, 0x07, uint8(value)}}
			}
			pPercentage = percentage
		}
	}
}

func expressionPercentage(expression int) int {
	var percent int
	if expression > 32293 {
		percent = 98
	} else if expression > 31003 {
		percent = 94
	} else if expression > 29713 {
		percent = 90
	} else if expression > 28199 {
		percent = 86
	} else if expression > 26684 {
		percent = 82
	} else if expression > 24749 {
		percent = 78
	} else if expression > 22814 {
		percent = 74
	} else if expression > 20808 {
		percent = 70
	} else if expression > 18803 {
		percent = 66
	} else if expression > 17179 {
		percent = 62
	} else if expression > 15555 {
		percent = 58
	} else if expression > 14227 {
		percent = 54
	} else if expression > 12889 {
		percent = 50
	} else if expression > 11469 {
		percent = 46
	} else if expression > 10040 {
		percent = 42
	} else if expression > 9019 {
		percent = 38
	} else if expression > 8178 {
		percent = 34
	} else if expression > 7180 {
		percent = 30
	} else if expression > 6183 {
		percent = 26
	} else if expression > 5450 {
		percent = 22
	} else if expression > 4718 {
		percent = 18
	} else if expression > 4316 {
		percent = 14
	} else if expression > 3915 {
		percent = 10
	} else {
		percent = 6
	}
	return percent
}
