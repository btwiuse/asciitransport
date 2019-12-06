// +build console

package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/containerd/console"
)

func main() {
	fmt.Println("Press ESC twice to exit.")
	current := console.Current()

	if err := current.SetRaw(); err != nil {
		panic(err)
	}

	go func() {
		for buf := make([]byte, 4096); ; {
			n, err := current.Read(buf)
			if err != nil {
				panic(err)
			}
			for _, conn := range ClientManager {
				_, err := conn.Write(buf[:n])
				if err != nil {
				}
			}
			str := fmt.Sprintf("%q", string(buf[:n]))
			log.Println(str)
			if str == `"\x1b\x1b"` {
				log.Println("BYE")
				current.Reset()
				os.Exit(0)
			}
		}
	}()

	log.Println("$ nc localhost 1337 # see what it actually looks like on a terminal")
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		panic(err)
	}

	for i := 0; ; i++ {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		ClientManager[i] = conn
	}
}

var ClientManager = make(map[int]net.Conn)
