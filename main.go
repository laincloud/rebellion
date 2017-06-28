package main

import (
	"flag"
	"fmt"

	"github.com/laincloud/rebellion/core"
	"github.com/mijia/sweb/log"
)

const (
	version = "2.3.1"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.Parse()
	fmt.Printf("Rebellion version: %s\n", version)
	if debug {
		fmt.Println("Running in debug mode")
		log.EnableDebug()
	}
	core.Run()
}
