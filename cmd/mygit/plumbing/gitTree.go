package plumbing

import (
	"bytes"
	"strings"
)

// Currently just object name
type GitTreeEntry string

type GitTree struct {
  Type GitObjectType
	Entries []GitTreeEntry
  hash string
}

func NewGitTree() *GitTree {
  return &GitTree{Type: TreeType}
}

// Deserialize implements GitObject. Only deserializes first level
func (t *GitTree) Deserialize(content []byte) {
  // while more data (can bound by provided size)
  toProcess := content
  for len(toProcess) > 0{
    endModeIndex := bytes.IndexByte(toProcess, ' ')
    // if endModeIndex == -1 {
    //   return errors.New("Malformed header, no space found to mark end of type property")
    // }

    endNameIndex := bytes.IndexByte(toProcess[endModeIndex+1:], '\x00')
    // if endNameIndex == -1 {
    //   return errors.New("Malformed header, no space found to mark end of type property")
    // }
    endNameIndex += endModeIndex + 1
    nameSegment := toProcess[endModeIndex+1:endNameIndex]

    t.Entries = append(t.Entries, GitTreeEntry(nameSegment))
  
    hashEndIndex := endNameIndex + 21

    toProcess = toProcess[hashEndIndex:]
  }
}

// GetHash implements GitObject.
func (t *GitTree) GetHash() string {
  if t.hash == "" {
    t.hash = hashObject(t)
  }
  return t.hash
}

// GetType implements GitObject.
func (t *GitTree) GetType() GitObjectType {
  return t.Type
}

// Serialize implements GitObject.
func (t *GitTree) Serialize() []byte {
	return []byte(t.PrettyPrint(NameOnly))
}

type TreeFormatOption string

const (
    NameOnly   TreeFormatOption = "nameOnly"
)

func (t *GitTree) PrettyPrint(format TreeFormatOption) string {
  result := new(strings.Builder)
  switch format {
  case NameOnly:
    for _, entry := range t.Entries {
      result.WriteString(string(entry))
      result.WriteString(string("\n"))
    } 
  default:
    panic("That tree formatting option is not implemented")
  }

  return result.String()
}

