package main

import (
	"CloudStorage/handlers"
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	store := cookie.NewStore([]byte("secretkey"))
	router.Use(sessions.Sessions("mysession", store))

	//----unsecure-ROUTES-------
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(302, "/dashboard")
	})

	router.GET("/registration", handlers.RenderReg)
	router.POST("/registration", handlers.RegLogic)

	router.GET("/login", handlers.RenderLogin)
	router.POST("/login", handlers.LoginLogic)

	router.GET("/logout", handlers.Logout)
	//----secure-ROUTES--------
	authorized := router.Group("/")
	authorized.Use(handlers.AuthRequired())
	{
		authorized.GET("/dashboard", handlers.RenderDashboard)
	}

	fmt.Println("Запуск сервера")
	router.Run(":9091")
}
