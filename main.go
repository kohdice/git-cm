package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version  = "v0.1.0"
	revision = "HEAD"
)

func main() {
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s (rev: %s)\n", version, revision)
		os.Exit(0)
	}

	os.Exit(doCommit())
}
