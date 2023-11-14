package main

import (
	"compress/zlib"
	"fmt"
	"io"
  "bytes"
	"os"
	"regexp"
  "errors"
  "flag"
)

func isValidSHA1(hash string) bool {
  // SHA1 hashes are 40 hex characters
	if len(hash) != 40 {
		return false
	}

	// Use a regular expression to check if the string is a valid hexadecimal
	match, _ := regexp.MatchString("^[a-fA-F0-9]{40}$", hash)
	return match
}

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

  var err error
	switch command := os.Args[1]; command {
	case "init":
    err = initGit()
  case "cat-file":
    err = catFile(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}

  if err != nil {
    os.Exit(1)
  }
}

func initGit() error {
  for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
    if err := os.MkdirAll(dir, 0755); err != nil {
      fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
      return err
    }
  }

  headFileContents := []byte("ref: refs/heads/master\n")
  if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
    fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
    return err
  }

  fmt.Println("Initialized git directory")
  return nil
}

func catFile(args []string) error {
  catFileCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)

  // Define flags for the cat-file subcommand
  pFlag := catFileCmd.Bool("p", false, "Some flag description")

  // Parse flags for the cat-file command
  err := catFileCmd.Parse(args)
  if err != nil {
      fmt.Fprintf(os.Stderr, "Error parsing flags for cat-file: %s\n", err)
      return err
  }

  // Check if the -p flag is provided
  if *pFlag {
    if catFileCmd.NArg() < 1 {
      fmt.Fprintf(os.Stderr, "usage: mygit cat-file -p {hash}\n")
      return err
    }

    hash := catFileCmd.Arg(0)
    if !isValidSHA1(hash) {
      fmt.Fprintf(os.Stderr, "The provided hash could not be verified, please provide a valid SHA1 hash\n")
      return err
    }

    dir, file := hash[:2], hash[2:]

    // open compressed blob 
    blobPath := fmt.Sprintf(".git/objects/%s/%s", dir, file)
    compressedBlob, err := os.Open(blobPath)
    if err != nil {
      if errors.Is(err, os.ErrNotExist) {
        fmt.Fprintf(os.Stderr, "The file corresponding to the hash %s does not exist.\n", hash)
      } else {
        fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
      }
      return err
    }   
    defer compressedBlob.Close()

    // set up blob decompression
    decompressedBlob, err := zlib.NewReader(compressedBlob)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error decompressing file: %s\n", err)
      return err
    }
    defer decompressedBlob.Close()

    // write decompressed blob to stdout
    blobData, err := io.ReadAll(decompressedBlob)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
      return err
    }

    // check for header & parse out file content
    nullCharIndex := bytes.IndexByte(blobData, '\x00')
    if nullCharIndex < 0 {
      fmt.Fprintf(os.Stderr, "Error reading file, no header found\n")
      return err
    }

    content := blobData[nullCharIndex+1:]

    fmt.Print(string(content))
  }
  return nil
}
