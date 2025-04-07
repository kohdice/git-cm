package main

import "fmt"

var (
	version  = "v0.0.1"
	revision = "HEAD"
)

func main() {
	fmt.Printf("%s (rev %s)", version, revision)
}
