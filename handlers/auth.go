package handlers

import (
	"CloudStorage/models"
	"CloudStorage/repository"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user")
		if user == nil {
			ctx.Redirect(302, "/login")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

func RenderReg(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "registration.html", gin.H{
		"title": "Вход",
	})
}

func RegLogic(ctx *gin.Context) {
	var form models.UserPass
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.HTML(http.StatusBadRequest, "registration.html", gin.H{
			"error": "Введите пароль и логин",
		})
		return
	}
	fmt.Println("Получен логин и пароль:", form.Username, form.Password)
	if repository.CheckUser(form) == true {
		ctx.HTML(http.StatusBadRequest, "registration.html", gin.H{
			"error": "Пользователь с таким именем уже существует",
		})
		return
	}
	repository.AddUser(form)

	session := sessions.Default(ctx)
	session.Set("user", form.Username)
	session.Save()
	ctx.Redirect(http.StatusSeeOther, "/dashboard")
}

func RenderLogin(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Вход",
	})
}

func LoginLogic(ctx *gin.Context) {
	var form models.UserPass
	if err := ctx.ShouldBind(&form); err != nil {
		ctx.HTML(http.StatusBadRequest, "login.html", gin.H{
			"error": "Введите пароль и логин",
		})
		return
	}
	fmt.Println("Попытка входа:", form.Username, form.Password)
	if repository.LoginUser(form) == true {
		session := sessions.Default(ctx)
		session.Set("user", form.Username)
		session.Save()
		ctx.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}

	RenderLogin(ctx)
}

func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Delete("user")
	session.Save()
	ctx.Redirect(http.StatusSeeOther, "/login")
}
