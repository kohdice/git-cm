package main

import (
	"errors"
	"fmt"
	"os"
)

// errQuit is a sentinel error returned when the user selects Quit.
var errQuit = errors.New("quit selected")

// exitWithError prints the error message to stderr and returns a non-zero status code.
func exitWithError(err error) int {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 1
}
