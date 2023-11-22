package plumbing

import (
	"compress/zlib"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

  object, err := readObject(hash)
  if err != nil {
    return err
  }

  if *pFlag {
    data, err := object.Serialize()
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error serializing object file: %s\n", err)
      return err
    }
    fmt.Print(string(data))
  }
    return nil
}

func readObject(hash string) (*GitObject, error) {
  dir, file := hash[:2], hash[2:]
  objectPath := filepath.Join(".git/objects", dir, file)

  // open compressed file 
  compressedObject, err := os.Open(objectPath)
  if err != nil {
    if errors.Is(err, os.ErrNotExist) {
      fmt.Fprintf(os.Stderr, "The object corresponding to the hash %s does not exist.\n", hash)
    } else {
      fmt.Fprintf(os.Stderr, "Error opening file: %s\n", err)
    }
    return nil, err
  }   
  defer compressedObject.Close()

  // setup decompressor
  decompressedObject, err := zlib.NewReader(compressedObject)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error decompressing file: %s\n", err)
    return nil, err
  }
  defer decompressedObject.Close()

  // read the file data
  // consider using bufio vs io.ReadAll if processing larger files
  objectData, err := io.ReadAll(decompressedObject)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
    return nil, err
  }

  object, err := DeserializeGitObject(objectData)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error deserializing object file: %s\n", err)
    return nil, err
  }

  return object, nil
}

