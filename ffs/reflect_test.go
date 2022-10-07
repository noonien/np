package ffs_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/noonien/np/ffs"
	"github.com/noonien/np/nptest"
)

func readNode(t *testing.T, v any, path string) ([]byte, error) {
	t.Helper()

	node, err := ffs.ToNode(v, &ffs.Params{Name: "/"})
	if err != nil {
		return nil, fmt.Errorf("to node: %w", err)
	}

	mnt, done := nptest.Serve(t, node)
	defer done()

	return os.ReadFile(mnt + path) //nolint:wrapcheck
}
