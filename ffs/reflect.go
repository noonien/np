package ffs

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/noonien/np"
)

var ErrCannotConvert = errors.New("cannot convert to Node")

// Params are default values that are used if a value does not implement Stat.
type Params struct {
	Name string
	Mode np.Mode
	Typ  uint16
	Dev  uint32
}

func (p *Params) fillStat(st *np.Stat) {
	if st.Name == "" {
		st.Name = p.Name
	}
	if p.Mode != 0 {
		st.Mode |= p.Mode
	}
	if st.Typ == 0 {
		st.Typ = p.Typ
	}
	if st.Dev == 0 {
		st.Dev = p.Dev
	}
}

// ToNode creates a np.Node from a value.
//
// The following conversions are available.
//
//	string -> file containing the string
//	[]byte -> file containing the data
//	[]T    -> directory of nodes converted from T
//	          if the a child node does not return a name, the array index is used
//	struct -> directory of fields tagged with `np:"name,opts"`
//	          available options are:
//	            write    - the node represented by this field should be writable
//	            exec    - the node represented by this field should be executable
//	            splat   - applied to a Dir Node, adds the children of the Dir to the struct
//	            omitnil - don't list the node if it's nil
//
//	          The following options mark fields to be used for Stat()
//	            stat  - Stat starts wth this value
//	            qid   - used as value for Stat.Qid
//	            name  - used as value for Stat.Name
//	            len   - used as value for Stat.Len
//	            typ   - used as value for Stat.Typ
//	            dev   - used as value for Stat.Dev
//	            ver   - used as value for Stat.Qid.Version
//	            atime - used as value for Stat.Atime
//	            mtime - used as value for Stat.Mtime
//	            uid   - used as value for Stat.Uid
//	            gid   - used as value for Stat.Gid
//	            muid  - used as value for Stat.Muid
func ToNode(v any, p *Params) (np.Node, error) {
	if p == nil {
		p = &Params{}
	}

	rv := reflect.ValueOf(v)

	switch v := v.(type) {
	case string:
		return newData(*p, []byte(v)), nil
	case []byte:
		return newData(*p, v), nil
	}

	var node np.Node
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
		node = reflectArray(rv, *p)
	case reflect.Struct:
		node = reflectStruct(rv, *p)
	default:
	}
	if node != nil {
		return node, nil
	}

	if n, ok := v.(np.Node); ok {
		return n, nil
	}

	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
		node, err := ToNode(rv.Interface(), p)
		if err == nil {
			return node, nil
		}
	}

	return nil, fmt.Errorf("%w: %T", ErrCannotConvert, v)
}

type paramWrap struct {
	params Params
	val    any
}

var (
	_ np.Node      = &paramWrap{}
	_ np.Unwrapper = &paramWrap{}
)

func (pw *paramWrap) Stat() (np.Stat, error) {
	var st np.Stat

	// if val is already a Node, use Stat from it
	// this happens when a struct field is a Node
	// and is wrapped because the Name is empty, or
	// Mode has to be set
	if node, ok := pw.val.(np.Node); ok {
		var err error
		st, err = node.Stat()
		if err != nil {
			return st, fmt.Errorf("stat: %w", err)
		}
	}

	if l, ok := pw.val.(interface{ Len() int }); ok {
		st.Length = uint64(l.Len())
	}
	pw.params.fillStat(&st)
	return st, nil
}

func (pw *paramWrap) Unwrap() any {
	return pw.val
}
