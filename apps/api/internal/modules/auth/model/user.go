package model

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateUserRequest struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Avatar string    `json:"avatar,omitempty"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar,omitempty"`
}
