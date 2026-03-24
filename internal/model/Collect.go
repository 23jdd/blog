package model

import "time"

type Collect struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	ArticleID    int       `db:"article_id"`
	ArticleTitle string    `db:"article_title"`
	AuthorID     int       `db:"author_id"`
	CreateTime   time.Time `db:"create_time"`
	UpdateTime   time.Time `db:"update_time"`
}
