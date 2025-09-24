package model

import "github.com/golang-jwt/jwt/v5"

// User 结构体
type User struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	City     string `json:"city"`
	Password string `json:"password,omitempty"`
}

// Product 结构体
type Product struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

// Claims JWT 的载荷结构体
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
