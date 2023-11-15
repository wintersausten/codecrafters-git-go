package main

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
	"regexp"
	"strconv"
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
  case "hash-object":
    err = hashObject(os.Args[2:])
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

var catFileCmd = flag.NewFlagSet("cat-file", flag.ExitOnError)
var pFlag = catFileCmd.Bool("p", false, "Pretty print based on object type")

func catFile(args []string) error {
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

  objectData, err := readObject(hash)
  if err != nil {
    return err
  }

  if *pFlag {
    // && assumed object type is blob
    err := printBlob(objectData)
    if err != nil {
      return err
    }
  }
    return nil
}

func printBlob (objectData []byte) error {
  
  // parse out header
  // TODO: move this into the object read (also create an object abstraction)
  nullCharIndex := bytes.IndexByte(objectData, '\x00')
  if nullCharIndex < 0 {
    fmt.Fprintf(os.Stderr, "Error reading file, no header found\n")
    return errors.New("No header found in object file")
  }

  content := objectData[nullCharIndex+1:]

  fmt.Print(string(content))
  return nil
}

func readObject(hash string) ([]byte, error) {
  dir, file := hash[:2], hash[2:]
  // TODO: use filepath
  objectPath := fmt.Sprintf(".git/objects/%s/%s", dir, file)

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
  
  return objectData, nil
}

var hashObjectCmd = flag.NewFlagSet("hashObject", flag.ExitOnError)
var wFlag = hashObjectCmd.Bool("w", false, "Write the object")

func hashObject(args []string) error {
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
    compressedObjectWriter.Write(object)
  }

  fmt.Print(hash)
  return nil
}

