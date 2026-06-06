package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func RegisterRoutes(router *gin.Engine, minioClient *minio.Client) {
	router.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(302, "/dashboard")
	})

	router.GET("/registration", RenderReg)
	router.POST("/registration", RegLogic)
	router.GET("/login", RenderLogin)
	router.POST("/login", LoginLogic)
	router.GET("/logout", Logout)

	authorized := router.Group("/")
	authorized.Use(AuthRequired())
	{
		authorized.GET("/dashboard", RenderDashboard)
		authorized.POST("/upload", func(ctx *gin.Context) {
			UploadFile(ctx, minioClient)
		})
		authorized.GET("/download/:fileID", func(ctx *gin.Context) {
			DownloadFile(ctx, minioClient)
		})
		authorized.GET("/preview/:fileID", func(ctx *gin.Context) {
			PreviewFile(ctx, minioClient)
		})
		authorized.POST("/rename-file/:fileID", RenameFile)
		authorized.POST("/create-folder", CreateFolder)
		authorized.GET("/folder/:folderID", OpenFolder)
		authorized.POST("/delete-file/:fileID", func(ctx *gin.Context) {
			DeleteFile(ctx, minioClient)
		})
		authorized.POST("/delete-folder/:folderID", DeleteFolder)
		authorized.POST("/rename-folder/:folderID", RenameFolder)
	}
}
