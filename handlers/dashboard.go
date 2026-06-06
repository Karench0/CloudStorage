package handlers

import (
	"CloudStorage/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func RenderDashboard(ctx *gin.Context) {
	username, userID, ok := currentUser(ctx)
	if !ok {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	files, directories, _ := loadRootBrowse(userID)

	renderBrowse(ctx, gin.H{
		"title":       "Личный кабинет",
		"Username":    username,
		"Files":       files,
		"Directories": directories,
		"CurrentDir":  nil,
	})
}

func CreateFolder(ctx *gin.Context) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	folderName := ctx.PostForm("folderName")
	if folderName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Имя папки не может быть пустым"})
		return
	}

	var parentID *int
	if parentStr := ctx.PostForm("parentID"); parentStr != "" {
		pid, err := strconv.Atoi(parentStr)
		if err == nil {
			parent, err := repository.GetDirectoryByID(pid)
			if err == nil && parent.UserID == userID {
				parentID = &pid
			}
		}
	}

	if err := repository.CreateDirectory(folderName, userID, parentID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка при создании папки"})
		return
	}

	redirectBrowse(ctx, parentID)
}

func OpenFolder(ctx *gin.Context) {
	username, userID, ok := currentUser(ctx)
	if !ok {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	folderID, err := strconv.Atoi(ctx.Param("folderID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID папки"})
		return
	}

	folder, files, subdirs, breadcrumbs, err := loadFolderBrowse(userID, folderID)
	if err != nil || folder.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой папке"})
		return
	}

	renderBrowse(ctx, gin.H{
		"title":       folder.Name,
		"Username":    username,
		"Files":       files,
		"Directories": subdirs,
		"CurrentDir":  folder,
		"Breadcrumbs": breadcrumbs,
	})
}

func DeleteFile(ctx *gin.Context, minioClient *minio.Client) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	fileID, err := strconv.Atoi(ctx.Param("fileID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID файла"})
		return
	}

	file, err := repository.GetFileByID(fileID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		return
	}

	if file.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этому файлу"})
		return
	}

	if err := repository.DeleteFileFromDB(fileID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении файла"})
		return
	}

	objectKey := file.Path
	if objectKey == "" {
		objectKey = file.Name
	}
	_ = repository.DeleteObject(minioClient, objectKey)

	ctx.JSON(http.StatusOK, gin.H{"message": "Файл успешно удален"})
}

func DeleteFolder(ctx *gin.Context) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	folderID, err := strconv.Atoi(ctx.Param("folderID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID папки"})
		return
	}

	folder, err := repository.GetDirectoryByID(folderID)
	if err != nil || folder.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой папке"})
		return
	}

	if err := repository.DeleteDirectory(folderID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении папки"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Папка успешно удалена"})
}

func RenameFolder(ctx *gin.Context) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	folderID, err := strconv.Atoi(ctx.Param("folderID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID папки"})
		return
	}

	newName := ctx.PostForm("newName")
	if newName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Новое имя не может быть пустым"})
		return
	}

	folder, err := repository.GetDirectoryByID(folderID)
	if err != nil || folder.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой папке"})
		return
	}

	if err := repository.RenameDirectory(folderID, newName); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при переименовании папки"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Папка успешно переименована"})
}
