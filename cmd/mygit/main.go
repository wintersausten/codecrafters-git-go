package main

import (
	"compress/zlib"
	"fmt"
	"os"
	"regexp"
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

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
        return
			}
		}

		headFileContents := []byte("ref: refs/heads/master\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
      return
		}

		fmt.Println("Initialized git directory")

  case "cat-file":
    // verify & parse hash
    if len(os.Args) != 3 {
      fmt.Fprintf(os.Stderr, "usage: mygit cat-file <hash>\n")
      return
    }

    hash := os.Args[2]
    if !isValidSHA1(hash) {
      fmt.Fprintf(os.Stderr, "The provided hash could not be verified, please provide a valid SHA1 hash\n")
      return
    }

    dir, file := hash[:2], hash[2:]

    // read file in
    blobPath := fmt.Sprintf(".git/objects/%s/%s", dir, file)
    compressedBlob, err := os.Open(blobPath)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
      return
    }
    defer compressedBlob.Close()

    // decompress
    decompressedBlob, err := zlib.NewReader(compressedBlob)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error decompressing file: %s\n", err)
      return
    }
    defer decompressedBlob.Close()

    // read decompressed data
    // output data

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
