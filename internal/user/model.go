package user

import "time"

type User struct {
	ID        int64     `json:"id"         gorm:"primaryKey"`
	Username  string    `json:"username"   gorm:"uniqueIndex"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
