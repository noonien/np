package ffs_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimpleArray(t *testing.T) {
	t.Parallel()

	type Foo struct {
		Hello string `np:",exec"`
	}

	v := []Foo{
		{
			Hello: "hello world",
		},
	}

	data, err := readNode(t, v, "/0/Hello")
	require.Nil(t, err)
	require.Equal(t, v[0].Hello, string(data))
}
