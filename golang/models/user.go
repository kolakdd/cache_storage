package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type (
	User struct {
		ID           uuid.UUID  `json:"id" db:"id"`
		Login        string     `json:"login" db:"login"`
		HashPassword string     `json:"hashPassword" db:"hash_password"`
		CreatedAt    *time.Time `json:"createdAt" db:"created_at"`
		UpdatedAt    *time.Time `json:"updatedAt" db:"updated_at"`
	}

	RegisterUserRequest struct {
		Login string `json:"login"`
	}
)

func (u User) CreateCRUD() {}

func NewUserDB(login string, password string) User {
	return User{uuid.NewV4(), login, password, nil, nil}
}
