package model

import "time"

type Article struct {
	ID         int       `db:"id"`
	Title      string    `db:"title"`
	CreateTime time.Time `db:"create_time"`
	UpdateTime time.Time `db:"update_time"`
	Path       string    `db:"path"`
	AuthorID   int       `db:"author_id"`
}
