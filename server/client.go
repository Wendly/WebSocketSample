package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/YangChing/WebSocketSample/server/model"
)


type Client interface {
	Send(post model.Post) Client
	GetUserName() string
	SetClientManager(manager ClientManager) Client
}

func NewClient(conn *websocket.Conn, userName string) Client {
	return &BaseClient{conn: conn, send: make(chan model.Post), userName: userName}
}

type BaseClient struct {
	conn     *websocket.Conn
	manager  ClientManager
	send     chan model.Post
	userName string
}

func (c *BaseClient) handleClient() {
	defer func() {
		c.manager.Logout(c)
		c.conn.Close()
		close(c.send)
	}()

	for {
		var post model.Post
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

func (c *BaseClient) Send(post model.Post) Client {
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
