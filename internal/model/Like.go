package model

import "time"

// Like 点赞模型
type Like struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	ArticleID  int       `db:"article_id" json:"article_id"`
	CreateTime time.Time `db:"create_time" json:"create_time"`
}
