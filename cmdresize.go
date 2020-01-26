package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/containerd/console"
	Pty "github.com/creack/pty"
)

func main() {
	pty, tty, err := Pty.Open()
	if err != nil {
		panic(err)
	}

	current := console.Current()
	defer current.Reset()
	if err := current.SetRaw(); err != nil {
		panic(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	// var cmd *exec.Cmd = exec.CommandContext(ctx, "bash", "-c", "[[ -t 0 ]] && echo 0 isatty || echo 0 isnotatty")
	cmd := exec.CommandContext(ctx, "bash")
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    int(tty.Fd()),
	}

	go io.Copy(pty, current)
	go io.Copy(current, pty)

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
			Pty.Setsize(tty, &Pty.Winsize{Rows: currentSize.Height, Cols: currentSize.Width})
			<-ch
		}
	}()

	log.Println(cmd.Run())
}
