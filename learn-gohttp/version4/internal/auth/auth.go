package auth

import (
	"go-learning/learn-gohttp/version4/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_secret_key_change_it_in_production")

// GenerateJWT 为指定用户生成一个新的JWT
func GenerateJWT(user *model.User) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &model.Claims{
		Username: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidatePassword 验证提供的密码是否与哈希匹配
func ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GetJWTKey 返回JWT密钥 (用于中间件)
func GetJWTKey() []byte {
	return jwtKey
}
