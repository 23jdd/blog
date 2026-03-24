package sql

import (
	"blog/internal/model"
	"database/sql"
	"encoding/json"

	"blog/internal/redis"
	"context"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type UserMapper struct {
	db *sqlx.DB
}

func NewUserMapper() *UserMapper {
	return &UserMapper{db: db}
}

// 初始化数据库连接
func init() {
	var err error
	db, err = sqlx.Connect("mysql", "root:1234@tcp(127.0.0.1:3306)/data?parseTime=true")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)

	// 启动时自动确保核心业务表存在，避免联调阶段因缺表导致接口报错
	if err = createCoreTables(); err != nil {
		panic(fmt.Sprintf("create tables failed: %v", err))
	}
}

func createCoreTables() error {
	ddlList := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			image VARCHAR(512) NULL,
			age INT NULL,
			gender VARCHAR(32) NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS article (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content LONGTEXT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL,
			author_id INT NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'draft',
			category_id INT NOT NULL DEFAULT 0,
			tags VARCHAR(512) NOT NULL DEFAULT '',
			cover_url VARCHAR(512) NOT NULL DEFAULT '',
			INDEX idx_article_author_id (author_id),
			INDEX idx_article_status (status),
			INDEX idx_article_category_id (category_id),
			INDEX idx_article_create_time (create_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS review (
			id INT AUTO_INCREMENT PRIMARY KEY,
			article_id INT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL,
			content TEXT NOT NULL,
			author_id INT NOT NULL,
			is_direct TINYINT(1) NOT NULL DEFAULT 1,
			parent_id INT NOT NULL DEFAULT 0,
			status VARCHAR(32) NOT NULL DEFAULT 'pending',
			INDEX idx_review_article_id (article_id),
			INDEX idx_review_author_id (author_id),
			INDEX idx_review_parent_id (parent_id),
			INDEX idx_review_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS collect (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			article_id INT NOT NULL,
			article_title VARCHAR(255) NOT NULL,
			author_id INT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL,
			UNIQUE KEY uniq_collect_user_article (user_id, article_id),
			INDEX idx_collect_user_id (user_id),
			INDEX idx_collect_article_id (article_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS article_like (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			article_id INT NOT NULL,
			create_time DATETIME NOT NULL,
			UNIQUE KEY uniq_like_user_article (user_id, article_id),
			INDEX idx_like_article_id (article_id),
			INDEX idx_like_user_id (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		`CREATE TABLE IF NOT EXISTS draft (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content LONGTEXT NOT NULL,
			create_time DATETIME NOT NULL,
			update_time DATETIME NOT NULL,
			author_id INT NOT NULL,
			INDEX idx_draft_author_id (author_id),
			INDEX idx_draft_update_time (update_time)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}

	for _, ddl := range ddlList {
		if _, err := db.Exec(ddl); err != nil {
			return err
		}
	}
	return nil
}
func CreateUserTable() error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS user (id INT AUTO_INCREMENT PRIMARY KEY, username VARCHAR(255), password VARCHAR(255))")
	if err != nil {
		return err
	}
	return nil
}
func (u *UserMapper) FindByUsername(username string) (*model.User, error) {
	var user model.User
	fmt.Println("username: ", username)
	err := u.db.Get(&user, "SELECT * FROM user WHERE username = ?", username)
	if err != nil {
		fmt.Println("err: ", err)
		return nil, err
	}
	return &user, nil
}

func (u *UserMapper) UpdateUserInfo(id int, user model.User) error {
	_, err := u.db.Exec("UPDATE user SET age = ?, gender = ? WHERE id = ?", user.Age, user.Gender, id)
	if err != nil {
		return err
	}
	// 更新后清理缓存，避免脏读
	_ = redis.Client.Del(context.Background(), fmt.Sprintf("%d:info", id)).Err()
	return nil
}

func BuildNullableUserInfo(age string, gender string) (sql.NullInt32, sql.NullString) {
	var ageVal sql.NullInt32
	if age != "" {
		if v, err := strconv.Atoi(age); err == nil {
			ageVal = sql.NullInt32{Int32: int32(v), Valid: true}
		}
	}
	var genderVal sql.NullString
	if gender != "" {
		genderVal = sql.NullString{String: gender, Valid: true}
	}
	return ageVal, genderVal
}
func (u *UserMapper) FindByID(id int) (*model.User, error) {
	var user model.User
	err := u.db.Select(&user, "SELECT * FROM user WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserMapper) FindAll() ([]*model.User, error) {
	var users []*model.User
	err := u.db.Select(&users, "SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (u *UserMapper) Insert(user *model.User) (int64, error) {
	result, err := u.db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		return -1, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}
func (u *UserMapper) Update(user *model.User) error {
	_, err := u.db.Exec("UPDATE user SET username = ?, password = ? WHERE id = ?", user.Username, user.Password, user.ID)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserMapper) UpdateImage(id int, image string) error {
	_, err := u.db.Exec("UPDATE user SET image = ? WHERE id = ?", image, id)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserMapper) Delete(id int) error {
	_, err := u.db.Exec("DELETE FROM user WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserMapper) FindByUsernameAndPassword(username, password string) (*model.User, error) {
	var user model.User
	err := u.db.Get(&user, "SELECT * FROM user WHERE username = ? AND password = ?", username, password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserMapper) FindByUsernameOrPassword(username, password string) (*model.User, error) {
	var user model.User
	err := u.db.Select(&user, "SELECT * FROM user WHERE username = ? OR password = ?", username, password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserMapper) FindByUsernameOrID(username string, id int) (*model.User, error) {
	var user model.User
	err := u.db.Select(&user, "SELECT * FROM user WHERE username = ? OR id = ?", username, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (u *UserMapper) FindByUsernameAndID(username string, id int) (*model.User, error) {
	var user model.User
	err := u.db.Select(&user, "SELECT * FROM user WHERE username = ? AND id = ?", username, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserInfoByID 根据用户ID获取用户信息
// redis put {id}:info
func (u *UserMapper) GetUserInfoByID(id int) (*model.User, error) {
	var user model.User
	key := fmt.Sprintf("%d:info", id)
	value, err := redis.Client.Get(context.Background(), key).Result()
	if err != nil {
		// 缓存中不存在，从数据库查询
		err = u.db.Get(&user, "SELECT username, image, age, gender FROM user WHERE id = ?", id)
		if err != nil {
			return nil, err
		}
		// 序列化用户信息并存储到缓存中
		userJSON, err := json.Marshal(user)
		if err != nil {
			return nil, err
		}
		redis.Client.Set(context.Background(), key, userJSON, 0)
		return &user, nil
	}
	err = json.Unmarshal([]byte(value), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
