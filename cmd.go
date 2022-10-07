package np

import (
	"bytes"
	"io"
	"strings"

	"go.rbn.im/neinp/message"
)

type LineCmdRecv struct {
	Handler func(string) error
}

type lineCmdFD struct {
	lcr *LineCmdRecv
	buf bytes.Buffer
}

var (
	_ Opener      = &LineCmdRecv{}
	_ io.WriterAt = &lineCmdFD{}
)

func (lcr *LineCmdRecv) Open(mode message.OpenMode) (any, uint32, error) {
	return &lineCmdFD{lcr: lcr}, 0, nil
}

func (fd *lineCmdFD) WriteAt(p []byte, off int64) (int, error) {
	var n int
	for {
		idx := bytes.IndexByte(p, '\n')
		if idx < 0 {
			n += len(p)
			fd.buf.Write(p)
			break
		}

		cmd := p[:idx]
		if fd.buf.Len() > 0 {
			fd.buf.Write(p[:idx])
			cmd = fd.buf.Bytes()
		}

		scmd := string(cmd)
		scmd = strings.TrimSpace(scmd)

		if len(cmd) > 0 {
			err := fd.lcr.Handler(scmd)
			if err != nil {
				return n, err
			}
		}

		p = p[idx+1:]
		n += idx + 1
		fd.buf.Reset()
	}

	return n, nil
}
