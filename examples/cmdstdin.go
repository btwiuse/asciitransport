package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
	// "github.com/containerd/console"
	// Pty "github.com/creack/pty"
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	// var cmd *exec.Cmd = exec.CommandContext(ctx, "bash", "-c", "[[ -t 0 ]] && echo 0 isatty || echo 0 isnotatty")
	cmd := exec.CommandContext(ctx, "bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
			Ctty:    int(os.Stdin.Fd()),
		}
	*/

	log.Println(cmd.Run())
}
