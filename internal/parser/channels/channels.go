package channels

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/jrcichra/wfh-organist/internal/common"
	"github.com/jrcichra/wfh-organist/internal/types"
)

func ReadFile(filename string) []types.MidiCSVRecord {
	f, err := os.Open(filename)
	common.Must(err)
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	common.Must(err)

	//parse the data
	var csvRecords []types.MidiCSVRecord
	for i, line := range data {
		if i > 0 {
			var rec types.MidiCSVRecord
			for j, field := range line {
				if j == 0 {
					s, err := strconv.ParseUint(field, 10, 32)
					common.Must(err)
					rec.InputChannel = uint8(s) - 1
				}
				if j == 1 {
					s, err := strconv.ParseUint(field, 10, 32)
					common.Must(err)
					rec.OutputChannel = uint8(s) - 1

				}
				if j == 2 {
					s, err := strconv.ParseInt(field, 10, 32)
					common.Must(err)
					rec.Offset = int(s)
				}
			}
			csvRecords = append(csvRecords, rec)
		}
	}
	return csvRecords
}

func CheckChannel(channel uint8, csvRecords []types.MidiCSVRecord) uint8 {
	ret := channel
	for _, msg := range csvRecords {
		if msg.InputChannel == channel {
			ret = msg.OutputChannel
		}
	}
	return ret
}

func CheckOffset(channel uint8, note uint8, csvRecords []types.MidiCSVRecord) uint8 {
	ret := int(note)
	for _, msg := range csvRecords {
		if msg.InputChannel == channel {
			ret = ret + msg.Offset
		}
	}
	return uint8(ret)
}
