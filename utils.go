package main

import "time"

func getCurrentTs() int64 {
	return time.Now().UnixNano() / 1000000
}
