package main

import (
	"fmt"
	"os"
)

func exitWithError(err error) int {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	return 1
}
