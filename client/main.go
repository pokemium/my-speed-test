package main

import (
	"context"
	"runtime"
)

const api = ""

var (
	maxConnections = runtime.NumCPU()
	payloadSizes   = []int{
		1562500,  // 1.5625MB
		6250000,  // 6.25MB
		12500000, // 12.5MB
		26214400, // 25MB
	}
)

func main() {
	ctx := context.Background()
}
