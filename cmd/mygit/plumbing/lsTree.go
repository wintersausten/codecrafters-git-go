package plumbing

import (
	"flag"
	"fmt"
	"os"
)

var lsTreeCmd = flag.NewFlagSet("ls-tree", flag.ExitOnError)
var nameOnlyFlag = lsTreeCmd.Bool("name-only", false, "List only the file name of tree entries")
func WriteTree(args []string) error {
  err := lsTreeCmd.Parse(args)
  if err != nil {
      fmt.Fprintf(os.Stderr, "Error parsing flags for ls-tree: %s\n", err)
      return err
  }

  if lsTreeCmd.NArg() != 1 {
    fmt.Fprintf(os.Stderr, "usage: mygit ls-tree {flags} {hash}\n")
    return err
  }

  hash := lsTreeCmd.Arg(0)
  if !isValidSHA1(hash) {
    fmt.Fprintf(os.Stderr, "The provided hash could not be verified, please provide a valid SHA1 hash\n")
    return err
  }

  object, err := readGitObject(hash)
  if err != nil {
    return err
  }

  data := object.Serialize()

  if *nameOnlyFlag {
    if err != nil {
      fmt.Fprintf(os.Stderr, "Error serializing object file: %s\n", err)
      return err
    }
    fmt.Print(string(data))
  }
  
  return nil
}
