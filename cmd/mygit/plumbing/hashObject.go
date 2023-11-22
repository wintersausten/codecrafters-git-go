package plumbing

import (
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
  object := NewGitObject(BlobType, fileContents, len(fileContents))
  objectFileContents := object.GetObjectFileContents()

  hash := generateSHA1(objectFileContents)

  // if w flag, Write
  if *wFlag {
    writeObject(objectFileContents, hash)
  }

  fmt.Print(hash)
  return nil
}

func generateSHA1(data []byte) string {
  hasher := sha1.New()
  hasher.Write(data)
  return hex.EncodeToString(hasher.Sum(nil))
}

func writeObject (contents []byte, hash string) error {
  dir, file := hash[:2], hash[2:]
  path := filepath.Join(".git/objects", dir)

  if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
    if err := os.MkdirAll(path, 0755); err != nil {
      fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
      return err
    }
  }

  path = filepath.Join(path, file)
  objectFile, err := os.Create(path)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error creating file: %s", err)
    return err
  }
  defer objectFile.Close()

  compressedObjectWriter := zlib.NewWriter(objectFile)
  if err != nil {
    fmt.Fprintf(os.Stderr, "Error creating file: %s", err)
    return err
  }
  defer objectFile.Close()
  defer compressedObjectWriter.Close()
  compressedObjectWriter.Write(contents)

  return nil
}
