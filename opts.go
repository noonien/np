package np

import "time"

type Option func(*server)

func Msize(msize uint32) Option {
	return func(s *server) {
		s.msize = msize
	}
}

func Debug(flags DebugFlags) Option {
	return func(s *server) {
		s.debug |= flags
	}
}

type StatModifierFn func(path []string, st *Stat, qidonly bool) error

func StatModifier(sm StatModifierFn) Option {
	return func(s *server) { s.statMods = append(s.statMods, sm) }
}

func DefaultOwner(user, group string) Option {
	return StatModifier(func(path []string, st *Stat, qidonly bool) error {
		if !qidonly {
			if st.Uid == "" {
				st.Uid = user
			}

			if st.Gid == "" {
				st.Gid = group
			}

			if st.Muid == "" {
				st.Muid = user
			}
		}
		return nil
	})
}

func Umask(umask Mode) Option {
	umask = ^umask
	return StatModifier(func(path []string, st *Stat, qidonly bool) error {
		st.Mode &= umask
		return nil
	})
}

func DefaultNow() Option {
	return StatModifier(func(path []string, st *Stat, qidonly bool) error {
		if !qidonly {
			now := time.Now()
			if st.Atime.IsZero() {
				st.Atime = now
			}
			if st.Mtime.IsZero() {
				st.Mtime = now
			}
		}
		return nil
	})
}
