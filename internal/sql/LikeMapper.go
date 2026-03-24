package sql

import (
	"blog/internal/model"
	"time"
)

type LikeMapper struct{}

func NewLikeMapper() *LikeMapper {
	return &LikeMapper{}
}

func (l *LikeMapper) Create(like *model.Like) (int64, error) {
	now := time.Now()
	res, err := db.Exec("INSERT INTO article_like (user_id, article_id, create_time) VALUES (?, ?, ?)", like.UserID, like.ArticleID, now)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (l *LikeMapper) Exists(userID int, articleID int) (bool, error) {
	var cnt int
	err := db.Get(&cnt, "SELECT COUNT(*) FROM article_like WHERE user_id = ? AND article_id = ?", userID, articleID)
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (l *LikeMapper) Delete(userID int, articleID int) error {
	_, err := db.Exec("DELETE FROM article_like WHERE user_id = ? AND article_id = ?", userID, articleID)
	return err
}

func (l *LikeMapper) CountByArticle(articleID int) (int, error) {
	var cnt int
	err := db.Get(&cnt, "SELECT COUNT(*) FROM article_like WHERE article_id = ?", articleID)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}
