package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
	t.Run("connection timeout", func(t *testing.T) {
		client := NewTelnetClient("192.0.2.1:4242", time.Millisecond*100, nil, nil)
		start := time.Now()
		err := client.Connect()
		elapsed := time.Since(start)

		require.Error(t, err)
		require.True(t, elapsed < time.Millisecond*150, "should respect timeout")
		require.Contains(t, err.Error(), "timeout", "should contain timeout error")
	})

	t.Run("send to closed server", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)

		addr := l.Addr().String()
		errClose := l.Close()
		if errClose != nil {
			return
		}

		client := NewTelnetClient(addr, time.Second, io.NopCloser(bytes.NewBufferString("test")), &bytes.Buffer{})
		err = client.Connect()
		require.Error(t, err)
	})

	t.Run("server closes connection", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := bytes.NewBufferString("bye server\n")
			out := &bytes.Buffer{}
			client := NewTelnetClient(l.Addr().String(), time.Second*5, io.NopCloser(in), out)

			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			require.NoError(t, client.Send())

			err := client.Receive()
			require.NoError(t, err)
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			defer func() { require.NoError(t, conn.Close()) }()

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			require.NoError(t, err)
			require.Equal(t, "bye server\n", string(buf[:n]))
		}()
		wg.Wait()
	})

	t.Run("sigint handling", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		client := NewTelnetClient(l.Addr().String(), time.Second, os.Stdin, os.Stdout)
		require.NoError(t, client.Connect())

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT)

		go func() {
			time.Sleep(100 * time.Millisecond)
			err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			require.NoError(t, err)
		}()

		<-sigCh
		require.NoError(t, client.Close())
	})
}
