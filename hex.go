package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func hexToRawStruct(input string) (Raw, error) {
	bytes := make([]byte, 0)
	for _, token := range strings.Fields(input) {
		// each token must be size 2
		if len(token) != 2 {
			return Raw{}, fmt.Errorf("Token must be size 2")
		}
		hexToken, err := hex.DecodeString(token)
		cont(err)
		bytes = append(bytes, hexToken...)
	}
	return Raw{
		Time: time.Now(),
		Data: bytes,
	}, nil
}
