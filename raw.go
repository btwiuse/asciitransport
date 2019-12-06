// +build console

package main

import (
	"fmt"
	"time"

	"github.com/containerd/console"
)

func main() {
	fmt.Println("vim-go")
	current := console.Current()
	defer current.Reset()

	if err := current.SetRaw(); err != nil {
		panic(err)
	}
	ws, err := current.Size()
	if err != nil {
		panic(err)
	}
	current.Resize(ws)
	time.Sleep(time.Hour)
}
