package main

import (
	"context"
	"io"
	"log"
	"os/exec"
	"time"

	"github.com/containerd/console"
)

func main() {
	current, _, _ := console.NewPty()
	local := console.Current()
	defer local.Reset()
	if err := local.SetRaw(); err != nil {
		panic(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	var cmd *exec.Cmd = exec.CommandContext(ctx, "bash")
	cmd.Stdin = current
	cmd.Stdout = current
	cmd.Stderr = current

	/*
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
			Ctty:    int(current.Fd()),
		}
	*/

	go io.Copy(current, local)
	go io.Copy(local, current)

	log.Println(cmd.Run())
}
