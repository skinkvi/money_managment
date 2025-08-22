package user

import "time"

type User struct {
	ID       int64
	Username string
	Email    string
	PassHash string
	CreateAt time.Time
	UpdateAt time.Time
}
