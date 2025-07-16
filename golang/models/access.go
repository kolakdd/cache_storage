package models

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

type (
	UserXObject struct {
		UserID    uuid.UUID  `json:"id" db:"user_id"`
		ObjectID  uuid.UUID  `json:"ownerId" db:"object_id"`
		CreatedAt *time.Time `json:"createdAt" db:"created_at"`
		UpdatedAt *time.Time `json:"updatedAt" db:"updated_at"`
	}
)
