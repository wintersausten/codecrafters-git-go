package plumbing

type GitBlob struct {
  Type GitObjectType
  content []byte
  hash string
}

func NewGitBlob() *GitBlob {
  return &GitBlob{Type: BlobType}
}

func (b *GitBlob) Deserialize(content []byte) {
  b.content = content
}

func (b *GitBlob) Serialize() []byte {
  return b.content
}

func (b *GitBlob) GetType() GitObjectType {
  return b.Type
}

func (b *GitBlob) GetHash() string {
  if b.hash == "" {
    b.hash = hashObject(b)
  }
  return b.hash
}
