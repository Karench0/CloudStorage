package handlers

import (
	"CloudStorage/config"
	"CloudStorage/models"
	"CloudStorage/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	UserID   int    `json:"user_id"`
	jwt.RegisteredClaims
}

func generateToken(username string, userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// Middleware аутентификации
func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusFound, "/login")
			ctx.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.SetCookie("token", "", -1, "/", "", false, true)
			ctx.Redirect(http.StatusFound, "/login")
			ctx.Abort()
			return
		}

		ctx.Set("username", claims.Username)
		ctx.Set("userID", claims.UserID)
		ctx.Next()
	}
}

func renderAuthPage(ctx *gin.Context, template string) {
	ctx.HTML(http.StatusOK, template, gin.H{
		"title": "Вход",
	})
}

func renderAuthError(ctx *gin.Context, template string, message string) {
	ctx.HTML(http.StatusBadRequest, template, gin.H{
		"error": message,
	})
}

func RenderReg(ctx *gin.Context) {
	renderAuthPage(ctx, "registration.html")
}

func RegLogic(ctx *gin.Context) {
	var form models.UserPass
	if err := ctx.ShouldBind(&form); err != nil {
		renderAuthError(ctx, "registration.html", "Введите пароль и логин")
		return
	}

	if len(form.Password) < 6 {
		renderAuthError(ctx, "registration.html", "Пароль должен содержать минимум 6 символов")
		return
	}

	if repository.CheckUser(form) {
		renderAuthError(ctx, "registration.html", "Пользователь с таким именем уже существует")
		return
	}

	if err := repository.AddUser(form); err != nil {
		renderAuthError(ctx, "registration.html", "Ошибка регистрации")
		return
	}

	userID, err := repository.GetUserID(form.Username)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	token, err := generateToken(form.Username, userID)
	if err != nil {
		renderAuthError(ctx, "registration.html", "Ошибка создания токена сессии")
		return
	}

	ctx.SetCookie("token", token, 86400, "/", "", false, true) // HttpOnly кука на сутки
	ctx.Redirect(http.StatusSeeOther, "/dashboard")
}

func RenderLogin(ctx *gin.Context) {
	renderAuthPage(ctx, "login.html")
}

func LoginLogic(ctx *gin.Context) {
	var form models.UserPass
	if err := ctx.ShouldBind(&form); err != nil {
		renderAuthError(ctx, "login.html", "Введите пароль и логин")
		return
	}

	userID, success := repository.LoginUser(form)
	if success {
		token, err := generateToken(form.Username, userID)
		if err != nil {
			renderAuthError(ctx, "login.html", "Ошибка авторизации")
			return
		}
		ctx.SetCookie("token", token, 86400, "/", "", false, true)
		ctx.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	renderAuthError(ctx, "login.html", "Неверный логин или пароль")
}

func Logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "/", "", false, true)
	ctx.Redirect(http.StatusSeeOther, "/login")
}
