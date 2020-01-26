// +build resize

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/containerd/console"
)

func main() {
	current := console.Current()

	slave, name, _ := console.NewPty()
	slaveCh := make(chan console.WinSize)

	go func(){
		for {
			newSize := <-slaveCh
			log.Println(name, newSize)
			log.Println(slave.Size())
			slave.Resize(newSize)
			log.Println(slave.Size())
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	for {
		currentSize, err := current.Size()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Println("====================")
		log.Println("master:", currentSize)
		slaveCh <- currentSize
		<-ch
	}
}
