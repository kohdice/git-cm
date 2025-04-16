package main

import (
	"errors"
	"fmt"
)

// doCommit executes the commit process by obtaining the repository info,
// running the TUI via runTUI, constructing the commit message, and performing the commit.
// This function returns an int status code (0 on success, or an error code if an error occurs),
// but the status code handling is left to the caller.
func doCommit() int {
	root, err := findRepoRoot()
	if err != nil {
		return exitWithError(err)
	}

	repo, err := openRepo(root)
	if err != nil {
		return exitWithError(err)
	}

	author, err := getAuthorInfo(repo)
	if err != nil {
		return exitWithError(err)
	}

	msg, err := runTUI()
	if err != nil {
		if errors.Is(err, errQuit) {
			fmt.Println("Quit selected")
			return 0
		}
		return exitWithError(err)
	}

	hash, err := commitRepo(repo, *author, msg)
	if err != nil {
		return exitWithError(err)
	}

	fmt.Printf("Commit created: %s\n", hash)
	return 0
}
