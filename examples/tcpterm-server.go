package main

import (
	"io"
	"net"
	"os/exec"
	"syscall"

	"github.com/google/goterm/term"
)

func main() {
	ln, err := net.Listen("tcp", ":12345")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handle(conn)
	}
}

func handle(c net.Conn) {
	defer c.Close()
	/*
		f, err := c.(*net.TCPConn).File()
		if err != nil {
			log.Println(err)
		}
		defer func(){
			c.Close()
			f.Close()
			// log.Println("c.Close()")
			// log.Println("f.Close()")
		}()
		log.Println(f.Name())

		state, _ := term.Attr(f)
		state.Raw()
		if err := state.Set(f); err != nil {
			log.Println(err)
		}*/
	pty, _ := term.OpenPTY()

	go io.Copy(c, pty.Master)
	go io.Copy(pty.Master, c)

	cmd := exec.Command("bash")
	cmd.Stdin = pty.Slave
	cmd.Stdout = pty.Slave
	cmd.Stderr = pty.Slave
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		// Ctty:    int(pty.Slave.Fd()),
	}
	cmd.Run()
	// log.Println("cmd exit")
}
