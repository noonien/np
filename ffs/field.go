package ffs

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/noonien/np"
	"go.rbn.im/neinp/qid"
)

type field struct {
	params  Params
	omitnil bool
	idx     []int
}

func (f *field) Node(rv reflect.Value) (np.Node, error) {
	rf := rv.FieldByIndex(f.idx)
	if f.omitnil && rf.IsNil() {
		return nil, nil
	}

	iface := rf.Interface()
	node, err := ToNode(iface, &f.params)
	if err == nil {
		st, err := node.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat: %w", err)
		}

		if st.Name == "" || st.Mode == 0 {
			return &paramWrap{
				params: f.params,
				val:    node,
			}, nil
		}

		return node, nil
	}

	if !errors.Is(err, ErrCannotConvert) {
		return nil, err
	}

	return &paramWrap{
		params: f.params,
		val:    iface,
	}, nil
}

// special fields that are used to for np.Stat.
type specialFields struct {
	stat    []int
	qid     []int
	name    []int
	length  []int
	typ     []int
	dev     []int
	version []int
	atime   []int
	mtime   []int
	uid     []int
	gid     []int
	muid    []int
}

func (sf *specialFields) toStat(st *np.Stat, rv reflect.Value) {
	if nst, ok := getField[np.Stat](rv, sf.stat); ok {
		*st = nst
	}
	if qid, ok := getField[qid.Qid](rv, sf.qid); ok {
		st.Qid = qid
	}
	if name, ok := getField[string](rv, sf.name); ok {
		st.Name = name
	}
	if length, ok := getField[uint64](rv, sf.length); ok {
		st.Length = length
	}
	if typ, ok := getField[uint16](rv, sf.typ); ok {
		st.Typ = typ
	}
	if dev, ok := getField[uint32](rv, sf.dev); ok {
		st.Dev = dev
	}
	if version, ok := getField[uint32](rv, sf.version); ok {
		st.Qid.Version = version
	}
	if atime, ok := getField[time.Time](rv, sf.atime); ok {
		st.Atime = atime
	}
	if mtime, ok := getField[time.Time](rv, sf.mtime); ok {
		st.Mtime = mtime
	}
	if uid, ok := getField[string](rv, sf.uid); ok {
		st.Uid = uid
	}
	if gid, ok := getField[string](rv, sf.gid); ok {
		st.Gid = gid
	}
	if muid, ok := getField[string](rv, sf.muid); ok {
		st.Muid = muid
	}
}

func getField[T any](rv reflect.Value, idx []int) (T, bool) { //nolint:ireturn // false pozitive on generic parameter
	var zero T
	if len(idx) == 0 {
		return zero, false
	}

	val, ok := rv.FieldByIndex(idx).Interface().(T)
	if !ok {
		panic("invalid type")
	}

	return val, ok
}
