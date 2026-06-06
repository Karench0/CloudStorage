package main

import (
	"CloudStorage/config"
	"CloudStorage/handlers"
	"CloudStorage/repository"
	"fmt"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.SetFuncMap(template.FuncMap{
		"divideFloat": func(a int64, b int64) float64 {
			return float64(a) / float64(b)
		},
		"formatBytes":   handlers.FormatBytes,
		"isPreviewable": handlers.IsPreviewable,
		"previewKind":   handlers.PreviewKind,
		"lower":         strings.ToLower,
	})

	router.LoadHTMLGlob("templates/*")

	// Инициализируем БД (Пул соединений)
	repository.InitDB()

	minioClient := repository.InitMinio()
	handlers.RegisterRoutes(router, minioClient)

	fmt.Println("Запуск сервера на порту", config.AppPort)
	router.Run(":" + config.AppPort)
}
