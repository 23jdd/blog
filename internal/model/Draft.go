package model

import "time"

type Draft struct {
	ID         int       `db:"id" json:"id"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
	AuthorID   int       `db:"author_id" json:"author_id"`
}
