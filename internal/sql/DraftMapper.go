package sql

import (
	"blog/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type DraftMapper struct {
	DB *sqlx.DB
}

func NewDraftMapper() *DraftMapper {
	return &DraftMapper{DB: db}
}

func (d *DraftMapper) Create(draft *model.Draft) (int64, error) {
	now := time.Now()
	result, err := d.DB.Exec("INSERT INTO draft (title, content, create_time, update_time, author_id) VALUES (?, ?, ?, ?, ?)", draft.Title, draft.Content, now, now, draft.AuthorID)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

func (d *DraftMapper) Update(draft *model.Draft) error {
	_, err := d.DB.Exec("UPDATE draft SET title = ?, content = ?, update_time = ? WHERE id = ?", draft.Title, draft.Content, time.Now(), draft.ID)
	return err
}

func (d *DraftMapper) Delete(id int) error {
	_, err := d.DB.Exec("DELETE FROM draft WHERE id = ?", id)
	return err
}

func (d *DraftMapper) FindByID(id int) (*model.Draft, error) {
	var draft model.Draft
	err := d.DB.Get(&draft, "SELECT * FROM draft WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &draft, nil
}

func (d *DraftMapper) FindAll(authorID int) ([]*model.Draft, error) {
	var drafts []*model.Draft
	err := d.DB.Select(&drafts, "SELECT * FROM draft WHERE author_id = ?", authorID)
	if err != nil {
		return nil, err
	}
	return drafts, nil
}

func (d *DraftMapper) FindAllCount(authorID int) (int, error) {
	var count int
	err := d.DB.Get(&count, "SELECT COUNT(*) FROM draft WHERE author_id = ?", authorID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
