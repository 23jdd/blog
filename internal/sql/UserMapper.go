package sql

import (
	"blog/internal/model"
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
	db, err = sqlx.Connect("mysql", "root:123456@tcp(127.0.0.1:3306)/blog")
	if err != nil {
		panic(err)
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
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
	err := u.db.Select(&user, "SELECT * FROM user WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
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
func (u *UserMapper) Insert(user *model.User) error {
	_, err := u.db.Exec("INSERT INTO user (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		return err
	}
	return nil
}
func (u *UserMapper) Update(user *model.User) error {
	_, err := u.db.Exec("UPDATE user SET username = ?, password = ? WHERE id = ?", user.Username, user.Password, user.ID)
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
	err := u.db.Select(&user, "SELECT * FROM user WHERE username = ? AND password = ?", username, password)
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
