package model

import "time"

type Article struct {
	ID         int       `db:"id" json:"id"`
	Title      string    `db:"title" json:"title"`
	Content    string    `db:"content" json:"content"` //markdown格式 url
	CreateTime time.Time `db:"create_time" json:"create_time"`
	UpdateTime time.Time `db:"update_time" json:"update_time"`
	AuthorID   int       `db:"author_id" json:"author_id"`
	Status     string    `db:"status" json:"status"`           // draft/published/offline
	CategoryID int       `db:"category_id" json:"category_id"` // 分类ID
	Tags       string    `db:"tags" json:"tags"`               // 标签，逗号分隔
	CoverURL   string    `db:"cover_url" json:"cover_url"`     // 封面图URL
}
