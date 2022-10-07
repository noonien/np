package np

import (
	"hash/fnv"

	"go.rbn.im/neinp/qid"
)

// fillstat fills empty stat fields.
func (s *server) fillstat(st *Stat, qidonly bool, path ...string) error {
	if len(s.statMods) > 0 {
		for _, sm := range s.statMods {
			err := sm(path, st, qidonly)
			if err != nil {
				return err
			}
		}
	}

	if st.Qid.Type == 0 && st.IsDir() {
		st.Qid.Type = qid.TypeDir
	}

	if st.Qid.Path == 0 {
		h := fnv.New64a()
		for _, p := range path {
			h.Write([]byte(p))
		}
		st.Qid.Path = h.Sum64()
	}

	return nil
}
