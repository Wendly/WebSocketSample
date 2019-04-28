package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("ip:")

	scanner.Scan()
	ip := scanner.Text()
	if ip == "" {
		ip = "127.0.0.1:12345"
	}

	fmt.Println("Starting application...")
	fmt.Println(ip)

	manager := NewClientManager()
	go manager.HandleMessage()

	http.HandleFunc("/ws", func(res http.ResponseWriter, req *http.Request) {
		conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
		if err != nil {
			http.NotFound(res, req)
			return
		}
		manager.Login(NewClient(conn, req.Header.Get("username")))
	})

	err := http.ListenAndServe(ip, nil)
	if err != nil {
		fmt.Println(err)
	}
}
