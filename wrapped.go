package np

// Unwrapper, when impleemnted by a Node, will have Unwrap() called on it before
// checking if it implements any interfaces the server requires (i.e Open, io.ReaderAt, etc).
type Unwrapper interface {
	// Unwraps a Node to maybe expose another interface
	//
	// The returned value might be an Unwrapper itself
	Unwrap() any
}

// Wrapped provides simple wrapping of a value with a Node.
type Wrapped struct {
	Node
	Val any
}

func (w *Wrapped) Unwrap() any { return w.Val }

// UnwrapValue is used on values to unwrap them until the requested type is found
// or until there is nothing else to unwrap (in which case, the second return value will be false).
func UnwrapValue[T any](v any) (T, bool) { //nolint:ireturn // returning an interface is the point of this function
	var z T
	for v != nil {
		if i, ok := v.(T); ok {
			return i, ok
		}

		if u, ok := v.(Unwrapper); ok {
			v = u.Unwrap()
		} else {
			v = nil
		}
	}
	return z, false
}
