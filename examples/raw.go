// +build console

package main

import (
	"fmt"
	"log"

	"github.com/containerd/console"
)

func main() {
	fmt.Println("Press ESC twice to exit.")
	current := console.Current()
	defer current.Reset()

	if err := current.SetRaw(); err != nil {
		panic(err)
	}
	for buf := make([]byte, 4096); ; {
		n, err := current.Read(buf)
		if err != nil {
			panic(err)
		}
		str := fmt.Sprintf("%q", string(buf[:n]))
		log.Println(str)
		if str == `"\x1b\x1b"` {
			log.Println("BYE")
			break
		}
	}
}
