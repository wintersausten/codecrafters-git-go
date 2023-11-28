package plumbing

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

type GitObjectType string

const (
    BlobType   GitObjectType = "blob"
    TreeType   GitObjectType = "tree"
    CommitType GitObjectType = "commit"
    TagType    GitObjectType = "tag"
)

type GitObject interface {
  Deserialize([]byte)
  Serialize() []byte 
  GetType() GitObjectType
  GetHash() string
}

func NewGitObject (objectType GitObjectType, contents []byte) GitObject {
  var gitObject GitObject
  switch objectType {
  case BlobType:
    blob := NewGitBlob()
    blob.Deserialize(contents)
    gitObject = blob
  case TreeType:
    tree := NewGitTree()
    tree.Deserialize(contents)
    gitObject = tree
  default:
    panic("Attempted to deserialize a type of git object that hasn't been implemented yet")
  }
  return gitObject
}

func DeserializeGitObject (data []byte) (GitObject, error) {
  endTypeIndex := bytes.IndexByte(data, ' ')
  if endTypeIndex == -1 {
    return nil, errors.New("Malformed header, no space found to mark end of type property")
  }
  typeSegment := data[:endTypeIndex]
  objectType := GitObjectType(string(typeSegment))

  endSizeIndex := bytes.IndexByte(data[endTypeIndex+1:], '\x00')
  if endSizeIndex == -1 {
    return nil, errors.New("Malformed header, no null character found to mark end of size property")
  }
  endSizeIndex += endTypeIndex + 1

  sizeSegment := data[endTypeIndex+1:endSizeIndex]
  size, err := strconv.Atoi(string(sizeSegment))

  contents := data[endSizeIndex + 1:]

  if err != nil || size != len(contents) {
    return nil, errors.New("Size property of object header is malformed")
  }

  return NewGitObject(objectType, contents), nil
}

func readGitObject(hash string) (GitObject, error) {
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

func writeGitObject (o GitObject) error {
  hash := o.GetHash()
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
  compressedObjectWriter.Write(GetObjectFileContents(o))

  return nil
}

func GetObjectFileContents (o GitObject) ([]byte) {
  content := o.Serialize()
  var objectBuffer bytes.Buffer

  objectBuffer.WriteString(string(o.GetType()))
  objectBuffer.WriteByte(' ')
  objectBuffer.WriteString(strconv.Itoa(len(content)))
  objectBuffer.WriteByte(0)
  objectBuffer.Write(content)

  return objectBuffer.Bytes()
}

func hashObject (o GitObject) string {
  data := GetObjectFileContents(o)
  hasher := sha1.New()
  hasher.Write(data)
  return hex.EncodeToString(hasher.Sum(nil))
}

