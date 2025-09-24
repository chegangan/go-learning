package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-sql-driver/mysql"
	"your_project_module_name/internal/auth"
	"your_project_module_name/internal/model"
	"your_project_module_name/internal/store"
)

// AuthHandler 包含认证相关的处理器
type AuthHandler struct {
	UserStore *store.UserStore
}

// RegisterHandler 处理用户注册
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// ... 注册逻辑，但调用 h.UserStore.RegisterUser() ...
}

// LoginHandler 处理用户登录
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// ... 登录逻辑，调用 h.UserStore.GetUserForLogin() 和 auth.ValidatePassword(), auth.GenerateJWT() ...
}
