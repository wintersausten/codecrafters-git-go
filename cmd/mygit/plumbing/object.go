package plumbing

import (
	"bytes"
	"errors"
	"strconv"
)

type GitObjectType string

const (
    BlobType   GitObjectType = "blob"
    TreeType   GitObjectType = "tree"
    CommitType GitObjectType = "commit"
    TagType    GitObjectType = "tag"
)

type GitObject struct {
  objectType GitObjectType
  data []byte
  size int
}

func NewGitObject (t GitObjectType, data []byte, size int) *GitObject {
  return &GitObject{
    objectType: t,
    data: data,
    size: size,
  }
}

func (o *GitObject) Serialize () ([]byte, error) {
  switch o.objectType {
  case BlobType:
    return o.data, nil
  default:
    err := errors.New("Serialization for that type is not yet supported")
    return nil, err
  }
}

func (o *GitObject) GetObjectFileContents () ([]byte) {
  var objectBuffer bytes.Buffer
  objectBuffer.WriteString(string(o.objectType))
  objectBuffer.WriteByte(' ')
  objectBuffer.WriteString(strconv.Itoa(o.size))
  objectBuffer.WriteByte(0)
  objectBuffer.Write(o.data)
  return objectBuffer.Bytes()
}

func DeserializeGitObject (data []byte) (*GitObject, error) {
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
  if err != nil {
    return nil, errors.New("Size property of object header is malformed")
  }
  return NewGitObject(objectType, data[endSizeIndex+1:], size), nil
}
