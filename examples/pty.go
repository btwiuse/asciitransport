package main

import (
	"context"
	"io"
	"log"
	"bytes"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/lxc/lxd/shared/termios"
	"github.com/containerd/console"
	"github.com/creack/pty"
)

// inspired by Tomas Senart's video

// WrapWriterFunc implementes io.Writer
type WrapWriterFunc func(p []byte) (n int, err error)

func (wr WrapWriterFunc) Write(p []byte) (n int, err error) {
	return wr(p)
}

// WrapReaderFunc implementes io.Reader
type WrapReaderFunc func(p []byte) (n int, err error)

func (wr WrapReaderFunc) Read(p []byte) (n int, err error) {
	return wr(p)
}

func CopyWithContext(ctx context.Context, w io.Writer, r io.Reader) error {
	ww := WrapWriterFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return w.Write(p)
		}
	})
	rr := WrapReaderFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return r.Read(p)
		}
	})
	_, err := io.Copy(ww, rr)
	return err
}

func main() {
	current := console.Current()
	defer current.Reset()
	if err := current.SetRaw(); err != nil {
		panic(err)
	}

	htop, err := pty.Start(exec.Command("htop"))
	if err != nil {
		panic(err)
	}

	dstat, err := pty.Start(exec.Command("bash", "-c", "dstat | ts"))
	if err != nil {
		panic(err)
	}

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		for {
			currentSize, err := current.Size()
			if err != nil {
				log.Println(err)
				continue
			}

			// fmt.Println(currentSize)
			pty.Setsize(htop, &pty.Winsize{Rows: currentSize.Height, Cols: currentSize.Width})
			pty.Setsize(dstat, &pty.Winsize{Rows: currentSize.Height, Cols: currentSize.Width})
			<-ch
		}
	}()

	htopBuf := bytes.NewBuffer([]byte{})
	dstatBuf := bytes.NewBuffer([]byte{})

	for {
		{
			log.Println(current.Fd(), htop.Fd(), dstat.Fd())
			htopState, err := termios.GetState(int(htop.Fd()))
			if err != nil {
				panic(err)
			}
			termios.Restore(int(current.Fd()), htopState)

			current.Write(htopBuf.Bytes())
			ctx, cancelhtop := context.WithCancel(context.Background())
			go CopyWithContext(ctx, htop, current)
			go CopyWithContext(ctx, current, io.TeeReader(htop, htopBuf))
			time.Sleep(10 * time.Second)
			cancelhtop()
		}

		{
			dstatState, err := termios.GetState(int(dstat.Fd()))
			if err != nil {
				panic(err)
			}
			termios.Restore(int(current.Fd()), dstatState)

			current.Write(dstatBuf.Bytes())
			ctx, canceldstat := context.WithCancel(context.Background())
			go CopyWithContext(ctx, dstat, current)
			go CopyWithContext(ctx, current, io.TeeReader(dstat, dstatBuf))
			time.Sleep(10 * time.Second)
			canceldstat()
		}
	}
}
