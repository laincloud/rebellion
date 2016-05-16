package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/laincloud/rebellion/core"
	"github.com/mijia/sweb/log"
)

const (
	version        = "2.0.3"
	defaultConfDir = "/etc/hekaconf.d/"
)

func main() {
	var confDir string
	var debug bool
	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	flag.StringVar(&confDir, "d", "", "The path of Heka configuration directory")
	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.Parse()
	fmt.Printf("Rebellion version: %s\n", version)
	if hostName := os.Getenv("NODE_NAME"); hostName == "" {
		fmt.Println("Get hostname failed! Program aborted")
		os.Exit(1)
	} else {
		fmt.Printf("Current host name: %s\n", hostName)
		if confDir == "" {
			fmt.Printf("No Heka configuration directory specified. Use %s as default\n", confDir)
			confDir = defaultConfDir
		}
		if debug {
			fmt.Println("Running in debug mode")
			log.EnableDebug()
		}
		r := core.NewRebellion(hostName, confDir, base)
		r.ListenAndUpdate()
	}
}
