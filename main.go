package main

import (
	"flag"
	"fmt"
	"os"
)

var version = "v0.0.1"

func main() {
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	os.Exit(doCommit())
}
