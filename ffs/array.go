package ffs

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/noonien/np"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"
)

func reflectArray(rv reflect.Value, p Params) *arrayDir {
	if p.Name == "" {
		elem := rv.Type().Elem()
		for elem.Kind() == reflect.Pointer {
			elem = elem.Elem()
		}
		p.Name = elem.Name() + "s"
	}

	return &arrayDir{
		rv:     rv,
		params: p,
	}
}

type arrayDir struct {
	rv     reflect.Value
	params Params
}

var (
	_ np.Node = &arrayDir{}
	_ np.Dir  = &arrayDir{}
)

func (ad *arrayDir) Stat() (np.Stat, error) {
	var st np.Stat

	if n, ok := ad.rv.Interface().(np.Node); ok {
		var err error
		if st, err = n.Stat(); err != nil {
			return st, fmt.Errorf("stat: %w", err)
		}
	}

	if st.Mode == 0 {
		st.Mode = 0o555
	}

	ad.params.fillStat(&st)
	st.Mode |= stat.Dir
	st.Qid.Type = qid.TypeDir

	return st, nil
}

func (ad *arrayDir) Children() ([]np.Stat, error) {
	rv := ad.rv

	arrlen := rv.Len()
	cstats := make([]np.Stat, 0, arrlen)
	cnames := make(map[string]struct{}, len(cstats))

	for i := 0; i < arrlen; i++ {
		av := rv.Index(i)
		istr := strconv.Itoa(i)

		node, err := ToNode(av.Interface(), &Params{Name: istr})
		if err != nil {
			return nil, err
		}

		st, err := node.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat: %w", err)
		}

		if _, ok := cnames[st.Name]; ok {
			continue
		}
		cnames[st.Name] = struct{}{}

		cstats = append(cstats, st)
	}

	return cstats, nil
}

func (ad *arrayDir) Walk(name string) (np.Node, error) {
	rv := ad.rv
	arrlen := rv.Len()

	for i := 0; i < arrlen; i++ {
		av := rv.Index(i)
		istr := strconv.Itoa(i)

		node, err := ToNode(av.Interface(), &Params{Name: istr})
		if err != nil {
			return nil, err
		}

		st, err := node.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat: %w", err)
		}

		if st.Name == name {
			return node, nil
		}
	}

	return nil, np.ErrNotFound
}
