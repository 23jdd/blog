package model

import "time"

type Review struct {
	ID         int       `db:"id"`          // 评论ID
	ArticleID  int       `db:"article_id"`  // 评论文章ID
	CreateTime time.Time `db:"create_time"` // 创建时间
	UpdateTime time.Time `db:"update_time"` // 更新时间
	Content    string    `db:"content"`     // 评论内容
	AuthorID   int       `db:"author_id"`   // 评论作者ID
	IsDirect   bool      `db:"is_direct"`   //  是否直接评论
	ParentID   int       `db:"parent_id"`   //  父评论ID
}
