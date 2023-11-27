package plumbing

type GitTreeEntry string

type GitTree struct {
  GitObject
  Entries []GitTreeEntry 
}

