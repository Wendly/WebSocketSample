package main

import (
	"github.com/YangChing/WebSocketSample/server/model"
)

type ClientManager interface {
	Login(client Client) ClientManager
	Logout(client Client) ClientManager
	Send(client Client, post model.Post) ClientManager
	HandleMessage()
}

func NewClientManager() ClientManager {
	return &BaseClientManager{
		broadcaster: make(chan model.Post),
		register:    make(chan Client),
		unregister:  make(chan Client),
		clients:     make(map[Client]bool)}
}

type BaseClientManager struct {
	clients     map[Client]bool
	broadcaster chan model.Post
	register    chan Client
	unregister  chan Client
}

func (m *BaseClientManager) HandleMessage() {
	for {
		select {
		case client := <-m.register:
			m.onLogin(client)
		case client := <-m.unregister:
			m.onLogout(client)
		case post := <-m.broadcaster:
			m.broadcast(post, nil)
		}
	}
}

func (m *BaseClientManager) onLogin(client Client) {
	client.OnLogin(m)
	m.clients[client] = true
	m.broadcast(model.NewPost(client.GetUserName(), "entry room"), client)
}

func (m *BaseClientManager) onLogout(client Client) {
	if _, ok := m.clients[client]; ok {
		delete(m.clients, client)
		m.broadcast(model.NewPost(client.GetUserName(), "leave room"), client)
	}
}

func (m *BaseClientManager) broadcast(post model.Post, ignore Client) {
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

func (m *BaseClientManager) Send(client Client, post model.Post) ClientManager {
	m.broadcaster <- post
	return m
}

