package model

import (
	"database/sql"
)

type User struct {
	ID       int            `db:"id"`
	Username string         `db:"username"`
	Password string         `db:"password"`
	Image    sql.NullString `db:"image"`  // 头像url
	Age      sql.NullInt32  `db:"age"`    // 年龄
	Gender   sql.NullString `db:"gender"` // 性别
}
