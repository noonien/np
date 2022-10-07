package ffs_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleStruct(t *testing.T) {
	t.Parallel()

	v := struct {
		Hello string `np:",exec"`
	}{
		Hello: "hello world",
	}

	data, err := readNode(t, v, "/Hello")
	require.Nil(t, err)
	require.Equal(t, v.Hello, string(data))
}
