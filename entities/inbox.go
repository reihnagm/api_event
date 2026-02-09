package entities

import "time"

type InboxList struct {
	Id         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Field1     string    `json:"field_1"`
	Field2     string    `json:"field_2"`
	Field3     string    `json:"field_3"`
	Field4     string    `json:"field_4"`
	Field5     string    `json:"field_5"`
	Field6     string    `json:"field_6"`
	Field7     string    `json:"field_7"`
	Field8     string    `json:"field_8"`
	Data       string    `json:"data"`
	Type       string    `json:"type"`
	ReceiverId string    `json:"receiver_id"`
	Status     string    `json:"status"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type InboxDetail struct {
	Id        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Field1    string    `json:"field_1"`
	Field2    string    `json:"field_2"`
	Field3    string    `json:"field_3"`
	Field4    string    `json:"field_4"`
	Field5    string    `json:"field_5"`
	Field6    string    `json:"field_6"`
	Field7    string    `json:"field_7"`
	Field8    string    `json:"field_8"`
	Data      string    `json:"data"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type InboxStore struct {
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	UserId     string         `json:"user_id"`
	ReceiverId string         `json:"receiver_id"`
	Field1     string         `json:"field_1"`
	Field2     string         `json:"field_2"`
	Field3     string         `json:"field_3"`
	Field4     string         `json:"field_4"`
	Field5     string         `json:"field_5"`
	Field6     string         `json:"field_6"`
	Data       map[string]any `json:"data"`
}

type Inboxes struct {
	ID         int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Title      string `gorm:"column:title" json:"title"`
	Content    string `gorm:"column:content" json:"content"`
	Field1     string `gorm:"column:field_1" json:"field_1"`
	Field2     int    `gorm:"column:field_2" json:"field_2"`
	Field3     string `gorm:"column:field_3" json:"field_3"`
	Field4     string `gorm:"column:field_4" json:"field_4"`
	Field5     string `gorm:"column:field_5" json:"field_5"`
	Field6     string `gorm:"column:field_6" json:"field_6"`
	Field7     int    `gorm:"column:field_7" json:"field_7"`
	Field8     string `gorm:"column:field_8" json:"field_8"`
	UserId     string `gorm:"column:user_id" json:"user_id"`
	ReceiverId string `gorm:"column:receiver_id" json:"receiver_id"`
	Type       string `gorm:"column:type" json:"type"`
}
