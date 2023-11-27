package plumbing

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var hashObjectCmd = flag.NewFlagSet("hashObject", flag.ExitOnError)
var wFlag = hashObjectCmd.Bool("w", false, "Write the object")

func HashObject(args []string) error {
  err := hashObjectCmd.Parse(args)
  if err != nil {
      fmt.Fprintf(os.Stderr, "Error parsing flags for hash-object: %s\n", err)
      return err
  }

  if hashObjectCmd.NArg() != 1 {
    fmt.Fprintf(os.Stderr, "usage: mygit hash-object {flags} {file}\n")
    return err
  }
  fileName := hashObjectCmd.Arg(0)

  // read the file in to mem
  file, err := os.Open(fileName)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error opening file.")
    return err
  }
  // TODO: replace with bufio to handle larger files
  fileContents, err := io.ReadAll(file)

  // create GitObject from file contents
  object := NewGitObject(BlobType, fileContents)
  hash := object.GetHash()

  // if w flag, Write
  if *wFlag {
    writeGitObject(object)
  }

  fmt.Print(hash)
  return nil
}
