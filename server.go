package np

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"go.rbn.im/neinp/fid"
	"go.rbn.im/neinp/message"
)

type DebugFlags uint32

const (
	DebugReceived DebugFlags = 0x01
	DebugSent     DebugFlags = 0x02
	DebugFlush    DebugFlags = 0x04
	DebugMessages DebugFlags = DebugReceived | DebugSent | DebugFlush
	DebugData     DebugFlags = 0x08

	DebugKnownErrors   DebugFlags = 0x10
	DebugUnknownErrors DebugFlags = 0x20
	DebugErrors        DebugFlags = DebugKnownErrors | DebugUnknownErrors

	DebugAll DebugFlags = DebugMessages | DebugErrors
)

type server struct {
	root Node

	msize    uint32
	statMods []StatModifierFn
	debug    DebugFlags

	fids *fid.Map

	tagsMu sync.RWMutex
	tags   map[uint16]context.CancelFunc
}

type response struct {
	cancelled *bool
	message.Message
}

// Serve starts a 9p server over the provided io.ReadWriter that serves the root Node.
// The root Node should be a Dir.
func Serve(ctx context.Context, rwc io.ReadWriteCloser, root Node, opts ...Option) error {
	defer rwc.Close()

	const DefaultMsize = 0x2000
	s := &server{
		root:  root,
		msize: DefaultMsize,
		fids:  fid.New(),
		tags:  map[uint16]context.CancelFunc{},
	}

	for _, opt := range opts {
		opt(s)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	in, rcvErr := s.rcv(ctx, rwc)
	out := s.handle(ctx, in)
	sendErr := s.send(ctx, out, rwc)

	select {
	case err := <-rcvErr:
		return fmt.Errorf("receive: %w", err)
	case err := <-sendErr:
		return fmt.Errorf("send: %w", err)
	case <-ctx.Done():
		return ctx.Err() //nolint:wrapcheck
	}
}

func (s *server) rcv(ctx context.Context, r io.Reader) (<-chan message.Message, <-chan error) {
	in := make(chan message.Message)
	errch := make(chan error, 1)
	done := ctx.Done()

	go func() {
		defer close(in)

		for {
			lr := io.LimitReader(r, int64(s.msize))

			var req message.Message
			_, err := req.Decode(lr)
			if err != nil {
				errch <- fmt.Errorf("9p decode: %w", err)
				return
			}

			if s.debug&DebugReceived != 0 {
				c := req.Content
				if s.debug&DebugData == 0 {
					if w, ok := c.(*message.TWrite); ok {
						nw := *w
						nw.Data = nil
						c = &nw
					}
				}
				log.Printf("<- %x %#v", req.Tag, c)
			}

			select {
			case <-done:
				errch <- ctx.Err()
			case in <- req:
			}
		}
	}()

	return in, errch
}

func (s *server) handle(ctx context.Context, in <-chan message.Message) <-chan response {
	out := make(chan response)
	done := ctx.Done()

	go func() {
		defer close(out)

		for {
			var req message.Message
			select {
			case req = <-in:
			case <-done:
				return
			}

			var cancelled bool
			rctx, cancel := context.WithCancel(ctx)

			s.tagsMu.Lock()
			s.tags[req.Tag] = func() {
				cancelled = true
				cancel()
			}
			s.tagsMu.Unlock()

			// TODO: limit number of in-flight requests

			go func() {
				c, err := s.process(req.Content)
				if err != nil {
					c = s.mapErr(err, req)
				}

				res := response{
					cancelled: &cancelled,
					Message: message.Message{
						Tag:     req.Tag,
						Content: c,
					},
				}

				select {
				case out <- res:
				case <-rctx.Done():
				case <-done:
					return
				}
			}()
		}
	}()

	return out
}

func (s *server) mapErr(err error, req message.Message) *message.RError {
	var reqfmt string
	if s.debug&DebugErrors != 0 { //nolint:nestif
		if s.debug&DebugReceived != 0 {
			reqfmt = fmt.Sprintf("tag:%x", req.Tag)
		} else {
			// if we're not debugging received messages, so we need to print the
			// message that caused the error
			c := req.Content
			if s.debug&DebugData == 0 {
				if w, ok := c.(*message.TWrite); ok {
					nw := *w
					nw.Data = nil
					c = &nw
				}
			}
			reqfmt = fmt.Sprintf("tag:%x %#v", req.Tag, c)
		}
	}

	var ne Error
	if ok := errors.As(err, &ne); ok {
		if s.debug&DebugKnownErrors != 0 {
			log.Printf("error: %s (req %s)", ne.err, reqfmt)
		}
		return &message.RError{Ename: ne.err}
	}

	if s.debug&DebugUnknownErrors != 0 {
		log.Printf("unknown error: %s (req %s)", err.Error(), reqfmt)
	}

	// map some errors?
	return &message.RError{Ename: ErrIO.err}
}

var ErrUnexpectedMessageType = errors.New("unexpected message type")

func (s *server) process(c message.Content) (message.Content, error) { //nolint:ireturn
	switch c := c.(type) {
	case *message.TVersion:
		return &message.RVersion{
			Msize:   s.msize,
			Version: "9P2000",
		}, nil

	case *message.TFlush:
		s.tagsMu.Lock()
		cancel, ok := s.tags[c.Oldtag]
		if ok {
			delete(s.tags, c.Oldtag)
			cancel()
		}
		s.tagsMu.Unlock()

		if s.debug&DebugFlush != 0 && ok {
			log.Printf("xx tag:%x", c.Oldtag)
		}

		return &message.RFlush{}, nil
	case *message.TAuth:
		return s.auth(*c)
	case *message.TAttach:
		return s.attach(*c)
	case *message.TWalk:
		return s.walk(*c)
	case *message.TOpen:
		return s.open(*c)
	case *message.TCreate:
		return s.create(*c)
	case *message.TRead:
		return s.read(*c)
	case *message.TWrite:
		return s.write(*c)
	case *message.TClunk:
		if err := s.clunk(*c); err != nil {
			return nil, err
		}
		return &message.RClunk{}, nil
	case *message.TRemove:
		return s.remove(*c)
	case *message.TStat:
		return s.stat(*c)
	case *message.TWstat:
		if err := s.wstat(*c); err != nil {
			return nil, err
		}
		return &message.RWstat{}, nil
	}

	panic(fmt.Sprintf("unexpected message type %W", c))
}

func (s *server) send(ctx context.Context, out <-chan response, w io.Writer) <-chan error {
	errch := make(chan error, 1)
	done := ctx.Done()

	go func() {
		for {
			var res response
			select {
			case res = <-out:
			case <-done:
				errch <- ctx.Err()
				return
			}

			s.tagsMu.Lock()
			cancel, ok := s.tags[res.Tag]
			if !ok || *res.cancelled {
				s.tagsMu.Unlock()
				continue
			}
			cancel()
			delete(s.tags, res.Tag)
			s.tagsMu.Unlock()

			if s.debug&DebugSent != 0 {
				c := res.Content
				if s.debug&DebugData == 0 {
					if r, ok := c.(*message.RRead); ok {
						nr := *r
						nr.Data = nil
						c = &nr
					}
				}
				log.Printf("-> %x %#v", res.Tag, c)
			}

			_, err := res.Encode(w)
			if err != nil {
				errch <- fmt.Errorf("9p encode: %w", err)
				return
			}
		}
	}()

	return errch
}
