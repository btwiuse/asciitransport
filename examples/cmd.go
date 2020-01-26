package main

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/containerd/console"
)

func main() {
	current := console.Current()
	defer current.Reset()
	if err := current.SetRaw(); err != nil {
		panic(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	var cmd *exec.Cmd = exec.CommandContext(ctx, "htop")
	cmd.Stdin = current
	cmd.Stdout = current
	cmd.Stderr = current

	log.Println(cmd.Run())
}
