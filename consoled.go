// +build console

package main

import (
	"io"
	"net"

	"github.com/containerd/console"
)

func main() {
	ln, err := net.Listen("tcp", ":1337")
	if err != nil {
		panic(err)
	}
	conn, err := ln.Accept()
	if err != nil {
		panic(err)
	}
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
	io.Copy(conn, current)
}
