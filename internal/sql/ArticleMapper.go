package sql

import (
	"blog/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type ArticleMapper struct {
	DB *sqlx.DB
}

func NewArticleMapper() *ArticleMapper {
	return &ArticleMapper{DB: db}
}
func (a *ArticleMapper) FindByID(id int) (*model.Article, error) {
	article := &model.Article{}
	err := a.DB.Get(article, "SELECT id, title, content, create_time, update_time, author_id, status, category_id, tags, cover_url FROM article WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return article, nil
}
func (a *ArticleMapper) FindAllLimit(limit int, offset int) ([]*model.Article, error) {
	var articles []*model.Article
	err := a.DB.Select(&articles, "SELECT id, title, content, create_time, update_time, author_id, status, category_id, tags, cover_url FROM article ORDER BY create_time DESC LIMIT ?,?", offset, limit)
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
	err := a.DB.Select(&articles, "SELECT id, title, content, create_time, update_time, author_id, status, category_id, tags, cover_url FROM article WHERE author_id = ? ORDER BY create_time DESC", authid)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func (a *ArticleMapper) FindByTag(tag string, limit int, offset int) ([]*model.Article, error) {
	var articles []*model.Article
	err := a.DB.Select(&articles, "SELECT id, title, content, create_time, update_time, author_id, status, category_id, tags, cover_url FROM article WHERE tags LIKE ? ORDER BY create_time DESC LIMIT ?,?", "%"+tag+"%", offset, limit)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func (a *ArticleMapper) FindByCategoryID(categoryID int, limit int, offset int) ([]*model.Article, error) {
	var articles []*model.Article
	err := a.DB.Select(&articles, "SELECT id, title, content, create_time, update_time, author_id, status, category_id, tags, cover_url FROM article WHERE category_id = ? ORDER BY create_time DESC LIMIT ?,?", categoryID, offset, limit)
	if err != nil {
		return nil, err
	}
	return articles, nil
}

func (a *ArticleMapper) Insert(article *model.Article) (int64, error) {
	now := time.Now()
	res, err := a.DB.Exec("INSERT INTO article (title, content, create_time, update_time, author_id, status, category_id, tags, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		article.Title, article.Content, now, now, article.AuthorID, article.Status, article.CategoryID, article.Tags, article.CoverURL)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func (a *ArticleMapper) UpdateByID(id int, article *model.Article) error {
	_, err := a.DB.Exec("UPDATE article SET title = ?, content = ?, category_id = ?, tags = ?, cover_url = ?, update_time = ? WHERE id = ?",
		article.Title, article.Content, article.CategoryID, article.Tags, article.CoverURL, time.Now(), id)
	return err
}

func (a *ArticleMapper) UpdateStatusByID(id int, status string) error {
	_, err := a.DB.Exec("UPDATE article SET status = ?, update_time = ? WHERE id = ?", status, time.Now(), id)
	return err
}

func (a *ArticleMapper) DeleteByID(id int) error {
	_, err := a.DB.Exec("DELETE FROM article WHERE id = ?", id)
	return err
}
func (a *ArticleMapper) SearchArticles(keyword string) ([]model.Article, error) {
	ats := []model.Article{}
	err := a.DB.Select(&ats, "SELECT id FROM article d where title like '%?%'", keyword)
	if err != nil {
		return nil, err
	}
	return ats, nil
}
