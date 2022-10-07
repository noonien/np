package np

import (
	"go.rbn.im/neinp/message"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"
)

type (
	OpenMode = message.OpenMode
	Qid      = qid.Qid
	Stat     = stat.Stat
	Mode     = stat.Mode
)

// Node represents a simple filesystem node, that can return a stat
//
// Types that implement this interface, can implement other interfaces to
// provide functionality:
//   - Reading: io.ReaderAt, io.ReadSeeker, io.Reader (only sequential reads are allowed)
//   - Writing: io.WriterAt, io.WriteSeeker, io.Writer (only sequential writes are allowed)
//   - Closing: io.Closer
//   - Opening: Opener
//   - Directories: Dir
type Node interface {
	Stat() (Stat, error)
}

// Dir allows listing and walking to children.
type Dir interface {
	Children() ([]Stat, error)
	Walk(name string) (Node, error)
}

// Opener allows Nodes to return new opened instances.
//
// If the returned val is not Node, it is Wrapped with the current Node.
type Opener interface {
	Open(OpenMode) (val any, iounit uint32, err error)
}
