package model

import (
	"database/sql"
	"time"
)

type User struct {
	ID       int            `db:"id" json:"id"`                 // 用户ID
	Username string         `db:"username" json:"username"`     // 用户名
	Password string         `db:"password" json:"password"`     // 密码
	Image    sql.NullString `db:"image" json:"image"`           // 头像url
	Age      sql.NullInt32  `db:"age" json:"age"`               // 年龄
	Gender   sql.NullString `db:"gender" json:"gender"`         // 性别
	CreateAt time.Time      `db:"created_at" json:"created_at"` // 创建时间
	UpdateAt time.Time      `db:"updated_at" json:"updated_at"` // 更新时间
}
