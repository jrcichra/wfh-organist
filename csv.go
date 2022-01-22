package main

import (
	"encoding/csv"
	"os"
	"strconv"
)

func readCSV() []MidiCSVRecord {
	f, err := os.Open("config.csv")
	must(err)
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	must(err)

	//parse the data
	var csvRecords []MidiCSVRecord
	for i, line := range data {
		if i > 0 {
			var rec MidiCSVRecord
			for j, field := range line {
				if j == 0 {
					s, err := strconv.ParseUint(field, 10, 32)
					must(err)
					rec.InputChannel = uint8(s) - 1
				}
				if j == 1 {
					s, err := strconv.ParseUint(field, 10, 32)
					must(err)
					rec.OutputChannel = uint8(s) - 1

				}
				if j == 2 {
					s, err := strconv.ParseInt(field, 10, 32)
					must(err)
					rec.Offset = int(s)
				}
			}
			csvRecords = append(csvRecords, rec)
		}
	}
	return csvRecords
}

func csvCheckChannel(channel uint8, csvRecords []MidiCSVRecord) uint8 {
	ret := channel
	for _, msg := range csvRecords {
		if msg.InputChannel == channel {
			ret = msg.OutputChannel
		}
	}
	// if ret == 255 {
	// 	log.Println("I'm blackholing channel", channel)
	// }
	return ret
}

func csvCheckOffset(channel uint8, note uint8, csvRecords []MidiCSVRecord) uint8 {
	ret := int(note)
	for _, msg := range csvRecords {
		if msg.InputChannel == channel {
			ret = ret + msg.Offset
		}
	}
	return uint8(ret)
}
