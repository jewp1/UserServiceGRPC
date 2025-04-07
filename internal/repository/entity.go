package repository

import "time"

type User struct {
	Email       string `validate:"required,email"`
	Username    string `validate:"required, max=1"`
	HashPass    string
	FirstName   string
	LastName    string
	Role        string
	LastLoginAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
