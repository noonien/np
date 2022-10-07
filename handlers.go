package np

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"go.rbn.im/neinp/message"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"
)

// walk starts from the root and walks the path from fd.path to get a Node.
func (s *server) walkfd(fd *fd) (Node, error) {
	node := s.root
	for _, name := range fd.path {
		dir, ok := UnwrapValue[Dir](node)
		if !ok {
			return nil, ErrWalkNoDir
		}

		var err error
		if node, err = dir.Walk(name); err != nil {
			return nil, fmt.Errorf("walk fd: %w", err)
		}
	}
	return node, nil
}

func (s *server) walk(m message.TWalk) (*message.RWalk, error) { //nolint:funlen
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return nil, ErrUnknownFid
	}

	if m.Fid != m.Newfid {
		if n := s.fids.Get(m.Newfid); n != nil {
			return nil, ErrDupFid
		}
	}

	node, err := s.walkfd(fd)
	if err != nil {
		return nil, ErrWalkNoDir
	}

	path := make([]string, 0, len(fd.path)+len(m.Wname))
	path = append(path, fd.path...)

	qids := make([]qid.Qid, 0, len(m.Wname))
	for i, name := range m.Wname {
		if name == ".." {
			pfd := fd.walk(m.Wname[:i]...).walk("..")
			if node, err = s.walkfd(pfd); err != nil {
				break
			}
		}

		dir, ok := UnwrapValue[Dir](node)
		if !ok {
			err = ErrWalkNoDir
			break
		}

		// TODO: permissions

		if node, err = dir.Walk(name); err != nil {
			break
		}

		var st Stat
		if st, err = node.Stat(); err != nil {
			break
		}

		path = append(path, name)
		if err = s.fillstat(&st, true, path...); err != nil {
			break
		}

		qids = append(qids, st.Qid)
	}

	if err != nil {
		if len(qids) > 0 {
			return &message.RWalk{Wqid: qids}, nil
		}
		return nil, err
	}

	nfd := fd.walk(m.Wname...)
	s.fids.Set(m.Newfid, nfd)
	return &message.RWalk{Wqid: qids}, nil
}

func (s *server) open(m message.TOpen) (*message.ROpen, error) {
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return nil, ErrUnknownFid
	}

	node, err := s.walkfd(fd)
	if fd == nil {
		return nil, err
	}

	// TODO: check permissions

	fd.mu.Lock()
	defer fd.mu.Unlock()

	var iounit uint32
	if o, ok := UnwrapValue[Opener](node); ok { //nolint:nestif
		var v any
		if v, iounit, err = o.Open(m.Mode); err != nil {
			return nil, fmt.Errorf("open: %w", err)
		}
		if vnode, ok := v.(Node); ok {
			node = vnode
		} else {
			node = &Wrapped{
				Node: node,
				Val:  v,
			}
		}
		fd.open = node
	} else if dir, ok := UnwrapValue[Dir](node); ok {
		dnode, ok := dir.(Node)
		if !ok {
			dnode = node
		}
		if node, err = s.newDir(dnode, dir, fd.path); err != nil {
			return nil, err
		}
		fd.open = node
	}

	var st Stat
	if st, err = node.Stat(); err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	if err = s.fillstat(&st, true, fd.path...); err != nil {
		return nil, err
	}

	return &message.ROpen{
		Qid:    st.Qid,
		Iounit: iounit,
	}, nil
}

func (s *server) create(m message.TCreate) (*message.RCreate, error) {
	return nil, ErrNoCreate
}

func (s *server) read(m message.TRead) (*message.RRead, error) {
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return nil, ErrUnknownFid
	}

	fd.mu.Lock()
	defer fd.mu.Unlock()

	var err error
	node := fd.open
	if node == nil {
		if node, err = s.walkfd(fd); err != nil {
			return nil, err
		}
	}

	// TODO: permissions

	var n int
	buf := make([]byte, m.Count)

	if ra, ok := UnwrapValue[io.ReaderAt](node); ok {
		// log.Printf("%v node is ReaderAt", fd.path)
		n, err = ra.ReadAt(buf, int64(m.Offset))
	} else if rs, ok := UnwrapValue[io.ReadSeeker](node); ok {
		// log.Printf("%v node is ReadSeeker", fd.path)
		if _, err = rs.Seek(int64(m.Offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek: %w", err)
		}
		n, err = rs.Read(buf)
	}

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read: %w", err)
	}

	return &message.RRead{Count: uint32(n), Data: buf[:n]}, nil
}

func (s *server) write(m message.TWrite) (*message.RWrite, error) {
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return nil, ErrUnknownFid
	}

	fd.mu.Lock()
	defer fd.mu.Unlock()

	var err error
	node := fd.open
	if node == nil {
		var err error
		if node, err = s.walkfd(fd); err != nil {
			return nil, err
		}
	}

	// TODO: permissions

	var n int
	if wa, ok := UnwrapValue[io.WriterAt](node); ok {
		n, err = wa.WriteAt(m.Data, int64(m.Offset))
	} else if ws, ok := UnwrapValue[io.WriteSeeker](node); ok {
		if _, err = ws.Seek(int64(m.Offset), io.SeekStart); err != nil {
			return nil, fmt.Errorf("seek: %w", err)
		}
		n, err = ws.Write(m.Data)
	} else {
		err = ErrNoWrite
	}

	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("write: %w", err)
	}

	return &message.RWrite{Count: uint32(n)}, nil
}

func (s *server) clunk(m message.TClunk) error {
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return ErrUnknownFid
	}
	s.fids.Delete(m.Fid)

	fd.mu.Lock()
	defer fd.mu.Unlock()

	var err error
	node := fd.open
	if node != nil {
		if node, err = s.walkfd(fd); err != nil {
			return err
		}
	}

	if c, ok := UnwrapValue[io.Closer](node); ok {
		err := c.Close()
		if err != nil {
			return fmt.Errorf("close: %w", err)
		}
	}

	return nil
}

func (s *server) remove(m message.TRemove) (*message.RRemove, error) {
	return nil, ErrNoRemove
}

func (s *server) stat(m message.TStat) (*message.RStat, error) {
	fd, _ := s.fids.Get(m.Fid).(*fd)
	if fd == nil {
		return nil, ErrUnknownFid
	}

	node, err := s.walkfd(fd)
	if fd == nil {
		return nil, err
	}

	st, err := node.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	if err = s.fillstat(&st, false, fd.path...); err != nil {
		return nil, err
	}

	return &message.RStat{
		Stat: st,
	}, nil
}

func (s *server) wstat(m message.TWstat) error {
	return nil
}

func (s *server) auth(m message.TAuth) (*message.RAuth, error) {
	return &message.RAuth{}, nil
}

func (s *server) attach(m message.TAttach) (*message.RAttach, error) {
	st, err := s.root.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	if err = s.fillstat(&st, true); err != nil {
		return nil, err
	}

	s.fids.Set(m.Fid, newfd())
	return &message.RAttach{
		Qid: st.Qid,
	}, nil
}

type dir struct {
	Node
	Dir
	*stat.Reader
}

func (s *server) newDir(n Node, d Dir, path []string) (*dir, error) {
	cs, err := d.Children()
	if err != nil {
		return nil, fmt.Errorf("children: %w", err)
	}
	for i := range cs {
		if err = s.fillstat(&cs[i], false, path...); err != nil {
			return nil, err
		}
	}

	r := stat.NewReader(cs...)
	return &dir{
		Node:   n,
		Dir:    d,
		Reader: r,
	}, nil
}

// func (s *server) close() error {}

type fd struct {
	path []string

	mu   sync.Mutex
	open Node
}

func newfd(path ...string) *fd {
	return &fd{path: path}
}

func (f *fd) walk(path ...string) *fd {
	if len(path) == 0 {
		return newfd(f.path...)
	}

	p := make([]string, 0, len(f.path)+len(path))
	p = append(p, f.path...)
	for _, name := range path {
		if name != ".." {
			p = append(p, name)
		} else if len(p) > 0 {
			p = p[:len(p)-1]
		}
	}
	return newfd(p...)
}
