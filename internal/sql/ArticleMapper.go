package sql

import (
	"blog/internal/model"

	"github.com/jmoiron/sqlx"
)

type ArticleMapper struct {
	DB *sqlx.DB
}

func NewArticleMapper(db *sqlx.DB) *ArticleMapper {
	return &ArticleMapper{DB: db}
}
func (a *ArticleMapper) FindByID(id int) (*model.Article, error) {
	article := &model.Article{}
	err := a.DB.Get(article, "SELECT  title, create_time, update_time, path, author_id FROM article WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return article, nil
}
func (a *ArticleMapper) FindAllLimit(limit int, offset int) ([]*model.Article, error) {
	var articles []*model.Article
	err := a.DB.Select(&articles, "SELECT id, title, create_time, update_time, path, author_id FROM article LIMIT ?,?", limit, offset)
	if err != nil {
		return nil, err
	}
	return articles, nil
}
func (a *ArticleMapper) FindAllCount() (int, error) {
	var count int
	err := a.DB.Get(&count, "SELECT COUNT(*) FROM article")
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (a *ArticleMapper) FindByAuthorID(authid int) ([]*model.Article, error) {
	var articles []*model.Article
	err := a.DB.Select(&articles, "SELECT id, title, create_time, update_time, path, author_id FROM article WHERE author_id = ?", authid)
	if err != nil {
		return nil, err
	}
	return articles, nil
}
func (a *ArticleMapper) Insert(article *model.Article) error {
	_, err := a.DB.Exec("INSERT INTO article (title, create_time, update_time, path, author_id) VALUES (?, ?, ?, ?, ?)",
		article.Title, article.CreateTime, article.UpdateTime, article.Path, article.AuthorID)
	if err != nil {
		return err
	}
	return nil
}
