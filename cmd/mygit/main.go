package main

import (
	"fmt"
	"os"

	"github.com/codecrafters-io/git-starter-go/cmd/mygit/plumbing"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

  var err error
	switch command := os.Args[1]; command {
	case "init":
    err = plumbing.InitRepo()
  case "cat-file":
    err = plumbing.CatFile(os.Args[2:])
  case "hash-object":
    err = plumbing.HashObject(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}

  if err != nil {
    os.Exit(1)
  }
}
