package main

import (
	"io"
	"log"
	"net/http"
	"os/exec"
	"syscall"

	"github.com/google/goterm/term"
	"github.com/gorilla/websocket"
	"github.com/btwiuse/wetty/pkg/utils"
)

func main() {
	upgrader := &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		rwc := utils.WsConnToReadWriter(conn)
		go handle(rwc)
	})
	log.Println(http.ListenAndServe("0.0.0.0:12345", nil))
}

func handle(c io.ReadWriteCloser) {
	defer c.Close()
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
	}
	cmd.Run()
}
