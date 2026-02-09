package entities

import "time"

type BroadcastList struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Path      string    `json:"path"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type BroadcastSend struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Path    string   `json:"path"`
	Date    string   `json:"date"`
	Cc      []string `json:"cc"`
}
