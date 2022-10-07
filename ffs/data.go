package ffs

import (
	"bytes"
	"io"

	"github.com/noonien/np"
)

type data struct {
	p Params
	*bytes.Reader
}

var (
	_ np.Node     = &data{}
	_ io.ReaderAt = &data{}
)

func newData(p Params, d []byte) *data {
	d = append([]byte{}, d...)
	return &data{
		p:      p,
		Reader: bytes.NewReader(d),
	}
}

func (d *data) Stat() (np.Stat, error) {
	var st np.Stat
	st.Length = uint64(d.Reader.Len())
	d.p.fillStat(&st)
	return st, nil
}
