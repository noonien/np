package nptest

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/noonien/np"
	"github.com/stretchr/testify/require"
)

// Serve mounts a Node at the returned mntpath, used for testing.
//
// The cleanup return function should be called at the end of the test to unmount and clean the files.
func Serve(t *testing.T, root np.Node, opts ...np.Option) (string, func()) {
	t.Helper()

	f, err := os.CreateTemp("", "np-test-sock")
	require.Nil(t, err)
	f.Close()
	unixpath := f.Name()
	os.Remove(unixpath)

	mntpath, err := os.MkdirTemp("", "np-test-mnt")
	require.Nil(t, err)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	cleanup := func() {
		cancel()
		wg.Wait()
	}

	l, err := net.Listen("unix", unixpath)
	require.Nil(t, err)

	go func() {
		<-ctx.Done()
		l.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer func() {
			exec.Command("9", "unmount", unixpath).Run() //nolint:errcheck
			os.Remove(unixpath)
			os.RemoveAll(mntpath)
		}()

		c, err := l.Accept()
		require.Nil(t, err)
		go func() {
			<-ctx.Done()
			c.Close()
		}()

		err = np.Serve(ctx, c, root, opts...)
		if !errors.Is(err, net.ErrClosed) && !errors.Is(err, context.Canceled) {
			require.Nil(t, err)
		}
	}()

	const MountWait = 10 * time.Millisecond
	time.Sleep(MountWait)

	err = exec.Command("9", "mount", "unix!"+unixpath, mntpath).Run() //nolint:gosec
	require.Nil(t, err)

	return mntpath, cleanup
}
