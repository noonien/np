package ffs

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/noonien/np"
	"github.com/ssoroka/slice"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"
)

var ErrSplatNode = errors.New("cannot splat non-DirNode")

func reflectStruct(rv reflect.Value, p Params) np.Node { //nolint:funlen,cyclop
	rt := rv.Type()
	rfs := reflect.VisibleFields(rt)

	// var params dirparams
	var fields []field
	var splats []field
	var special specialFields
	var hasSpecial bool
nextField:
	for _, rf := range rfs {
		if !rf.IsExported() {
			continue
		}

		t, ok := rf.Tag.Lookup("np")
		if !ok {
			tag := string(rf.Tag)
			tag = strings.TrimSpace(tag)
			if !strings.HasSuffix(tag, "np") {
				continue
			}
		}
		ts := slice.Map[string, string](strings.Split(t, ","), strings.TrimSpace)

		var fp Params
		var splat, omitnil bool
		fp.Mode = 0o444
		for _, p := range ts[1:] {
			switch p {
			case "write":
				fp.Mode |= 0o222
			case "exec":
				fp.Mode |= 0o111
			case "splat":
				splat = true
			case "omitnil":
				omitnil = true
			default:
				isSpecial := true
				switch p {
				case "stat":
					special.stat = rf.Index
				case "qid":
					special.qid = rf.Index
				case "name":
					special.name = rf.Index
				case "len":
					special.length = rf.Index
				case "typ":
					special.typ = rf.Index
				case "dev":
					special.dev = rf.Index
				case "ver":
					special.version = rf.Index
				case "atime":
					special.atime = rf.Index
				case "mtime":
					special.mtime = rf.Index
				case "uid":
					special.uid = rf.Index
				case "gid":
					special.gid = rf.Index
				case "muid":
					special.muid = rf.Index
				default:
					isSpecial = false
				}
				if isSpecial {
					hasSpecial = true
					continue nextField
				}
			}
		}

		fp.Name = ts[0]
		if fp.Name == "" {
			fp.Name = rf.Name
		}

		field := field{
			params:  fp,
			omitnil: omitnil,
			idx:     rf.Index,
		}

		if splat {
			splats = append(splats, field)
		} else {
			fields = append(fields, field)
		}
	}

	sn := structNode{
		rv:      rv,
		params:  p,
		special: special,
	}

	// this struct should present as a directory if it has any
	// fields that present as children, or splats
	if len(fields) > 0 || len(splats) > 0 {
		return &structDir{
			structNode: sn,
			fields:     fields,
			splats:     splats,
		}
	}

	iface := rv.Interface()

	// there are no children, but there are special fields
	// use them to wrap the value
	if hasSpecial {
		return &np.Wrapped{
			Node: &sn,
			Val:  iface,
		}
	}

	if n, ok := iface.(np.Node); ok {
		return n
	}

	return nil
}

type structNode struct {
	rv      reflect.Value
	params  Params
	special specialFields
}

var _ np.Node = &structNode{}

func (sn *structNode) Stat() (np.Stat, error) {
	var st np.Stat

	if n, ok := sn.rv.Interface().(np.Node); ok {
		var err error
		if st, err = n.Stat(); err != nil {
			return st, fmt.Errorf("stat: %w", err)
		}
	}

	if st.Mode == 0 {
		st.Mode = 0o555
	}

	sn.params.fillStat(&st)
	sn.special.toStat(&st, sn.rv)
	st.Qid.Type = qid.TypeDir
	st.Mode |= stat.Dir
	return st, nil
}

type structDir struct {
	structNode
	fields []field
	splats []field
}

var (
	_ np.Node = &structDir{}
	_ np.Dir  = &structDir{}
)

func (sd *structDir) Children() ([]np.Stat, error) {
	cstats := make([]np.Stat, 0, len(sd.fields)+len(sd.splats)*2)
	cnames := make(map[string]struct{}, len(cstats))

	for _, f := range sd.fields {
		node, err := f.Node(sd.rv)
		if err != nil {
			return nil, err
		}
		if node == nil {
			continue
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

	for _, s := range sd.splats {
		node, err := s.Node(sd.rv)
		if err != nil {
			return nil, err
		}
		if node == nil {
			continue
		}

		dnode, ok := np.UnwrapValue[np.Dir](node)
		if !ok {
			return nil, ErrSplatNode
		}

		children, err := dnode.Children()
		if err != nil {
			return nil, fmt.Errorf("children: %w", err)
		}

		for _, st := range children {
			if _, ok := cnames[st.Name]; ok {
				continue
			}
			cnames[st.Name] = struct{}{}
			cstats = append(cstats, st)
		}
	}

	sort.Slice(cstats, func(i, j int) bool {
		return cstats[i].Name < cstats[j].Name
	})

	return cstats, nil
}

func (sd *structDir) Walk(name string) (np.Node, error) {
	for _, f := range sd.fields {
		node, err := f.Node(sd.rv)
		if err != nil {
			return nil, err
		}
		if node == nil {
			continue
		}

		st, err := node.Stat()
		if err != nil {
			return nil, fmt.Errorf("stat: %w", err)
		}

		if st.Name == name {
			return node, nil
		}
	}

	for _, s := range sd.splats {
		node, err := s.Node(sd.rv)
		if err != nil {
			return nil, err
		}
		if node == nil {
			continue
		}

		dnode, ok := np.UnwrapValue[np.Dir](node)
		if !ok {
			return nil, ErrSplatNode
		}

		if node, err = dnode.Walk(name); err == nil {
			return node, nil
		}

		if errors.Is(err, np.ErrNotFound) {
			continue
		}

		return nil, fmt.Errorf("walk: %w", err)
	}

	return nil, np.ErrNotFound
}
