package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Post struct {
	UserName string `json:"username,omitempty"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

type Client interface {
	Send(post Post) Client
	GetUserName() string
	SetClientManager(manager ClientManager) Client
}

func NewClient(conn *websocket.Conn, userName string) Client {
	return &BaseClient{conn: conn, send: make(chan Post), userName: userName}
}

type BaseClient struct {
	conn     *websocket.Conn
	manager  ClientManager
	send     chan Post
	userName string
}

func (c *BaseClient) handleClient() {
	defer func() {
		c.manager.Logout(c)
		c.conn.Close()
		close(c.send)
	}()

	for {
		var post Post
		err := c.conn.ReadJSON(&post)
		if err != nil {
			fmt.Println("err", err)
			break
		}
		c.manager.Send(c, post)
	}
}

func (c *BaseClient) handleServer() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case post, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			message, _ := json.Marshal(&post)
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func (c *BaseClient) Send(post Post) Client {
	c.send <- post
	return c
}

func (c *BaseClient) GetUserName() string {
	return c.userName
}

func (c *BaseClient) SetClientManager(manager ClientManager) Client {
	c.manager = manager
	go c.handleClient()
	go c.handleServer()
	return c
}

type ClientManager interface {
	Login(client Client) ClientManager
	Logout(client Client) ClientManager
	Send(client Client, post Post) ClientManager
	HandleMessage()
}

func NewClientManager() ClientManager {
	return &BaseClientManager{
		broadcaster: make(chan Post),
		register:    make(chan Client),
		unregister:  make(chan Client),
		clients:     make(map[Client]bool)}
}

type BaseClientManager struct {
	clients     map[Client]bool
	broadcaster chan Post
	register    chan Client
	unregister  chan Client
}

func (m *BaseClientManager) HandleMessage() {
	for {
		select {
		case client := <-m.register:
			client.SetClientManager(m)
			m.clients[client] = true
			post := Post{UserName: client.GetUserName(), Time: time.Now().Format("2006-01-02 15:04:05"), Message: "entry room"}
			m.broadcast(post, client)
		case client := <-m.unregister:
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				post := Post{UserName: client.GetUserName(), Time: time.Now().Format("2006-01-02 15:04:05"), Message: "leave room"}
				m.broadcast(post, client)
			}
		case post := <-m.broadcaster:
			m.broadcast(post, nil)
		}
	}
}

func (m *BaseClientManager) broadcast(post Post, ignore Client) {
	for client := range m.clients {
		if client != ignore {
			client.Send(post)
		}
	}
}

func (m *BaseClientManager) Login(client Client) ClientManager {
	m.register <- client
	return m
}

func (m *BaseClientManager) Logout(client Client) ClientManager {
	m.unregister <- client
	return m
}

func (m *BaseClientManager) Send(client Client, post Post) ClientManager {
	m.broadcaster <- post
	return m
}

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
