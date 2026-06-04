package user

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) Store {
	return &store{db: db}
}

func (s *store) Create(ctx context.Context, user *User) error {
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *store) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).Where("username = ?", username).
		First(&user).Error; err != nil {
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &user, nil
}
