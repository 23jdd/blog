package sql

import (
	"blog/internal/model"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReviewMapper struct {
	DB *sqlx.DB
}

func NewReviewMapper(db *sqlx.DB) *ReviewMapper {
	return &ReviewMapper{DB: db}
}

func (r *ReviewMapper) Create(review *model.Review) (int64, error) {
	status := review.Status
	if status == "" {
		status = "pending"
	}
	result, err := r.DB.Exec(`
		INSERT INTO review (article_id, create_time, update_time, content, author_id, is_direct, parent_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		review.ArticleID,
		time.Now(),
		time.Now(),
		review.Content,
		review.AuthorID,
		review.IsDirect,
		review.ParentID,
		status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *ReviewMapper) FindByID(id int) (*model.Review, error) {
	review := &model.Review{}
	err := r.DB.Get(review, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return review, nil
}

func (r *ReviewMapper) FindByArticleID(articleID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE article_id = ? ORDER BY create_time DESC`, articleID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindByAuthorID(authorID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE author_id = ? ORDER BY create_time DESC`, authorID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindByParentID(parentID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE parent_id = ? ORDER BY create_time ASC`, parentID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindDirectReviews(articleID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE article_id = ? AND is_direct = 1 ORDER BY create_time DESC`, articleID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindReplies(parentID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE parent_id = ? AND is_direct = 0 ORDER BY create_time ASC`, parentID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindAll(limit int, offset int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review ORDER BY create_time DESC LIMIT ?, ?`, offset, limit)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindAllCount() (int, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review")
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewMapper) FindByArticleIDWithPagination(articleID int, limit int, offset int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE article_id = ? ORDER BY create_time DESC LIMIT ?, ?`, articleID, offset, limit)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindByArticleIDCount(articleID int) (int, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review WHERE article_id = ?", articleID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewMapper) Update(review *model.Review) error {
	_, err := r.DB.Exec(`
		UPDATE review 
		SET content = ?, update_time = ?
		WHERE id = ?`,
		review.Content,
		time.Now(),
		review.ID,
	)
	return err
}

func (r *ReviewMapper) UpdateContent(id int, content string) error {
	_, err := r.DB.Exec(`
		UPDATE review 
		SET content = ?, update_time = ?
		WHERE id = ?`,
		content,
		time.Now(),
		id,
	)
	return err
}

func (r *ReviewMapper) DeleteByID(id int) error {
	_, err := r.DB.Exec("DELETE FROM review WHERE id = ?", id)
	return err
}

func (r *ReviewMapper) DeleteByIDAndAuthor(id int, authorID int) error {
	_, err := r.DB.Exec("DELETE FROM review WHERE id = ? AND author_id = ?", id, authorID)
	return err
}

func (r *ReviewMapper) DeleteByArticleID(articleID int) error {
	_, err := r.DB.Exec("DELETE FROM review WHERE article_id = ?", articleID)
	return err
}

func (r *ReviewMapper) DeleteByAuthorID(authorID int) error {
	_, err := r.DB.Exec("DELETE FROM review WHERE author_id = ?", authorID)
	return err
}

func (r *ReviewMapper) DeleteByParentID(parentID int) error {
	_, err := r.DB.Exec("DELETE FROM review WHERE parent_id = ?", parentID)
	return err
}

func (r *ReviewMapper) ExistsByID(id int) (bool, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ReviewMapper) FindLatestByAuthorID(authorID int, limit int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE author_id = ? ORDER BY create_time DESC LIMIT ?`, authorID, limit)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindByArticleIDAndAuthorID(articleID int, authorID int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE article_id = ? AND author_id = ? ORDER BY create_time DESC`, articleID, authorID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) CountByArticleID(articleID int) (int, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review WHERE article_id = ?", articleID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewMapper) CountByAuthorID(authorID int) (int, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review WHERE author_id = ?", authorID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewMapper) CountReplies(parentID int) (int, error) {
	var count int
	err := r.DB.Get(&count, "SELECT COUNT(*) FROM review WHERE parent_id = ?", parentID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ReviewMapper) FindWithTimeRange(articleID int, startTime time.Time, endTime time.Time) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review WHERE article_id = ? AND create_time >= ? AND create_time <= ? 
		ORDER BY create_time DESC`, articleID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindRecentReviews(limit int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT id, article_id, create_time, update_time, content, author_id, is_direct, parent_id
		FROM review ORDER BY create_time DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) FindPopularReviews(limit int) ([]*model.Review, error) {
	var reviews []*model.Review
	err := r.DB.Select(&reviews, `
		SELECT r.id, r.article_id, r.create_time, r.update_time, r.content, r.author_id, r.is_direct, r.parent_id
		FROM review r
		LEFT JOIN (
			SELECT parent_id, COUNT(*) as reply_count 
			FROM review 
			WHERE parent_id IS NOT NULL 
			GROUP BY parent_id
		) reply_stats ON r.id = reply_stats.parent_id
		ORDER BY COALESCE(reply_stats.reply_count, 0) DESC, r.create_time DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *ReviewMapper) BatchCreate(reviews []*model.Review) (int64, error) {
	if len(reviews) == 0 {
		return 0, nil
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	query := `
		INSERT INTO review (article_id, create_time, update_time, content, author_id, is_direct, parent_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var lastID int64
	for _, review := range reviews {
		status := review.Status
		if status == "" {
			status = "pending"
		}
		result, err := stmt.Exec(
			review.ArticleID,
			now,
			now,
			review.Content,
			review.AuthorID,
			review.IsDirect,
			review.ParentID,
			status,
		)
		if err != nil {
			return 0, err
		}
		if lastID == 0 {
			lastID, _ = result.LastInsertId()
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (r *ReviewMapper) UpdateStatus(id int, status string) error {
	_, err := r.DB.Exec("UPDATE review SET status = ?, update_time = ? WHERE id = ?", status, time.Now(), id)
	return err
}
