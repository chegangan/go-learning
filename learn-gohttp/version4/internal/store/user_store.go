package store

import (
	"database/sql"
	"go-learning/learn-gohttp/version4/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// UserStore 包含与用户相关的数据库操作
type UserStore struct {
	db *sql.DB
}

// dbRegisterUser 在数据库中创建新用户，并哈希密码
func (s *UserStore) dbRegisterUser(u model.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (name, age, city, password_hash) VALUES (?, ?, ?, ?)", u.Name, u.Age, u.City, string(hashedPassword))
	return err
}

// dbGetUserForLogin 从数据库获取用户信息用于登录验证
func (s *UserStore) dbGetUserForLogin(name string) (*model.User, string, error) {
	var u model.User
	var passwordHash string
	row := s.db.QueryRow("SELECT name, age, city, password_hash FROM users WHERE name = ?", name)
	if err := row.Scan(&u.Name, &u.Age, &u.City, &passwordHash); err != nil {
		return nil, "", err
	}
	return &u, passwordHash, nil
}

// dbGetAllUsers 从数据库获取所有用户
func (s *UserStore) dbGetAllUsers() ([]model.User, error) {
	rows, err := s.db.Query("SELECT name, age, city FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.Name, &u.Age, &u.City); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// dbGetUserByName 从数据库按姓名获取单个用户
func (s *UserStore) dbGetUserByName(name string) (*model.User, error) {
	var u model.User
	row := s.db.QueryRow("SELECT name, age, city FROM users WHERE name = ?", name)
	if err := row.Scan(&u.Name, &u.Age, &u.City); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 没有找到用户，不是一个错误
		}
		return nil, err
	}
	return &u, nil
}
