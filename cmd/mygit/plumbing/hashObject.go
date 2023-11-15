package plumbing

import (
	"bytes"
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
	"strconv"
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

  // add header (assumes blob)
  var objectBuffer bytes.Buffer
  objectBuffer.WriteString("blob")
  objectBuffer.WriteByte(' ')
  objectBuffer.WriteString(strconv.Itoa(len(fileContents)))
  objectBuffer.WriteByte(0)
  objectBuffer.Write(fileContents)
  object := objectBuffer.Bytes()
  

  // hash file file contents 
  hasher := sha1.New()
  hasher.Write(object)
  hash := hex.EncodeToString(hasher.Sum(nil))

  // if w flag, Write
  if *wFlag {
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
    compressedObjectWriter.Write(object)
  }

  fmt.Print(hash)
  return nil
}


