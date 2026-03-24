package sql

import (
	"blog/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type CollectMapper struct {
	DB *sqlx.DB
}

func NewCollectMapper(db *sqlx.DB) *CollectMapper {
	return &CollectMapper{DB: db}
}
func (c *CollectMapper) Create(collect *model.Collect) (int64, error) {
	now := time.Now()
	result, err := c.DB.Exec("INSERT INTO collect (user_id, article_id, article_title, author_id, create_time, update_time) VALUES (?, ?, ?, ?, ?, ?)", collect.UserID, collect.ArticleID, collect.ArticleTitle, collect.AuthorID, now, now)
	if err != nil {
		return -1, err
	}
	collect.CreateTime = now
	collect.UpdateTime = now
	return result.LastInsertId()
}

// 统计收藏数
// 统计用户收藏数
func (c *CollectMapper) CountUserCollectByID(id int) (int, error) {
	var count int
	err := c.DB.Get(&count, "SELECT COUNT(*) FROM collect WHERE id = ?", id)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func NewCollectMapperDefault() *CollectMapper {
	return &CollectMapper{DB: db}
}

func (c *CollectMapper) Exists(userID int, articleID int) (bool, error) {
	var count int
	err := c.DB.Get(&count, "SELECT COUNT(*) FROM collect WHERE user_id = ? AND article_id = ?", userID, articleID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (c *CollectMapper) Delete(userID int, articleID int) error {
	_, err := c.DB.Exec("DELETE FROM collect WHERE user_id = ? AND article_id = ?", userID, articleID)
	return err
}

func (c *CollectMapper) ListByUser(userID int, limit int, offset int) ([]*model.Collect, error) {
	var rows []*model.Collect
	err := c.DB.Select(&rows, "SELECT id, user_id, article_id, article_title, author_id, create_time, update_time FROM collect WHERE user_id = ? ORDER BY create_time DESC LIMIT ?, ?", userID, offset, limit)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// 统计用户收藏包含作者数量
func (c *CollectMapper) CountUserCollectContainAuthor(id int) (int, error) {
	var count int
	err := c.DB.Get(&count, "SELECT COUNT(*) FROM collect WHERE id  = ? group by author_id", id)
	if err != nil {
		return 0, err
	}
	return count, nil
}
