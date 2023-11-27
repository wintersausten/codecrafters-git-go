package plumbing

import (
	"flag"
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
var catFileCmd = flag.NewFlagSet("cat-file", flag.ExitOnError)
var pFlag = catFileCmd.Bool("p", false, "Pretty print based on object type")

func CatFile(args []string) error {
  // Parse flags for the cat-file command
  err := catFileCmd.Parse(args)
  if err != nil {
      fmt.Fprintf(os.Stderr, "Error parsing flags for cat-file: %s\n", err)
      return err
  }

  if catFileCmd.NArg() != 1 {
    fmt.Fprintf(os.Stderr, "usage: mygit cat-file {flags} {hash}\n")
    return err
  }

  hash := catFileCmd.Arg(0)
  if !isValidSHA1(hash) {
    fmt.Fprintf(os.Stderr, "The provided hash could not be verified, please provide a valid SHA1 hash\n")
    return err
  }

  object, err := readGitObject(hash)
  if err != nil {
    return err
  }

  if *pFlag {
    data := object.Serialize()
    fmt.Print(string(data))
  }

  return nil
}
