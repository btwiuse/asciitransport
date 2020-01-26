// +build resize

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerd/console"
)

func main() {
	current := console.Current()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	for {
		currentSize, err := current.Size()
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println(currentSize)
		<-ch
	}
}
