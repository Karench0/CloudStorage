package handlers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func RenderDashboard(ctx *gin.Context) {
	session := sessions.Default(ctx)
	ctx.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "Личный кабинет",
		"Username": session.Get("user"),
	})
}
