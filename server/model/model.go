package model

import (
	"time"
)

type Post struct {
	UserName string `json:"username,omitempty"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

func NewPost(name string, message string) Post {
	return Post{name, "entry room", time.Now().Format("2006-01-02 15:04:05")}
}

