package model

import "time"

type Region string

const (
	IND Region = "IND"
	USA Region = "USA"
	CHN Region = "CHN"
	DEU Region = "DEU"
	SGP Region = "SGP"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	EmailAddress string    `json:"email_address"`
	Password     string    `json:"password"`
	Region       Region    `json:"region"`
	CreatedAt    time.Time `json:"created_at"`
}

type Post struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Content    string    `json:"content"`
	Region     Region    `json:"region"`
	CreatedAt  time.Time `json:"created_at"`
	AuthorName string    `json:"author_name"`
}

type Session struct {
	Token     string
	UserID    string
	ExpiresAt time.Time
}
