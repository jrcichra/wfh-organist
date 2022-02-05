package main

import (
	"time"
)

func handleMs(m time.Time) int64 {
	ms := time.Since(m).Milliseconds()
	return ms
}
