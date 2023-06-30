package main

import "time"

type User struct {
	ID           int       `json:"id"`
	UserName     string    `json:"user_name"`
	Password     string    `json:"password"`
	RefreshToken string    `json:"refresh_token"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}
