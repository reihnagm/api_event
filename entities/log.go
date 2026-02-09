package entities

import "time"

type LogList struct {
	Id        int       `json:"id"`
	Msg       string    `json:"msg"`
	Field1    string    `json:"field"1`
	CreatedAt time.Time `json:"created_at"`
}
