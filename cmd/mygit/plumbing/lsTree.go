package plumbing

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var lsTreeCmd = flag.NewFlagSet("ls-tree", flag.ExitOnError)
var nameOnlyFlag = lsTreeCmd.Bool("name-only", false, "List only the file name of tree entries")

func LsTree(args []string) error {
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

  tree, ok := object.(*GitTree);
  // assert object is a gittree
  if !ok {
    fmt.Fprintf(os.Stderr, "The object associated with the provided hash is not a tree\n")
    return errors.New("The object associated with the provided hash is not a tree\n")
  }

  var format TreeFormatOption
  if *nameOnlyFlag {
    format = NameOnly
  }

  fmt.Print(tree.PrettyPrint(format))
  
  return nil
}
