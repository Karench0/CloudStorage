package handlers

import (
	"CloudStorage/models"
	"CloudStorage/repository"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func UploadFile(ctx *gin.Context, minioClient *minio.Client) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	directoryID := parseDirectoryID(ctx, userID)

	if err := ctx.Request.ParseMultipartForm(64 << 20); err != nil {
		setUploadErrorAndRedirect(ctx, "Файл не выбран", directoryID)
		return
	}

	headers := ctx.Request.MultipartForm.File["file"]
	if len(headers) == 0 {
		setUploadErrorAndRedirect(ctx, "Файл не выбран", directoryID)
		return
	}

	uploaded := 0
	for _, header := range headers {
		if err := saveUploadedFile(minioClient, userID, directoryID, header); err != nil {
			fmt.Println("Ошибка загрузки:", err)
			if uploaded == 0 {
				setUploadErrorAndRedirect(ctx, "Не удалось загрузить файл. Проверьте, что MinIO запущен.", directoryID)
				return
			}
			break
		}
		uploaded++
	}

	redirectBrowse(ctx, directoryID)
}

func saveUploadedFile(minioClient *minio.Client, userID int, directoryID *int, header *multipart.FileHeader) error {
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	fileName := filepath.Base(header.Filename)
	if fileName == "" || fileName == "." {
		return fmt.Errorf("пустое имя файла")
	}

	fileSize := header.Size
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(fileName))
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectKey := repository.BuildObjectKey(userID, directoryID, fileName)
	if err := repository.PutObject(minioClient, objectKey, file, fileSize, contentType); err != nil {
		return err
	}

	fileObj := models.File{
		Name:        fileName,
		Size:        fileSize,
		Path:        objectKey,
		UserID:      userID,
		DirectoryID: directoryID,
		Extension:   filepath.Ext(fileName),
		ContentType: contentType,
	}

	if err := repository.SaveFile(fileObj); err != nil {
		_ = repository.DeleteObject(minioClient, objectKey)
		return err
	}
	return nil
}

func DownloadFile(ctx *gin.Context, minioClient *minio.Client) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	fileIDStr := ctx.Param("fileID")
	var fileID int
	if _, err := fmt.Sscanf(fileIDStr, "%d", &fileID); err != nil {
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

	objectKey := file.Path
	if objectKey == "" {
		objectKey = file.Name
	}

	object, err := repository.GetObject(minioClient, objectKey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении файла"})
		return
	}
	defer object.Close()

	contentType := mime.TypeByExtension(filepath.Ext(file.Name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	ctx.Header("Content-Type", contentType)
	if file.Size > 0 {
		ctx.Header("Content-Length", fmt.Sprintf("%d", file.Size))
	}

	_, _ = io.Copy(ctx.Writer, object)
}

func ownedFileObject(ctx *gin.Context, minioClient *minio.Client, fileID int) (*models.File, io.ReadCloser, error) {
	_, userID, ok := currentUser(ctx)
	if !ok {
		return nil, nil, fmt.Errorf("unauthorized")
	}

	file, err := repository.GetFileByID(fileID)
	if err != nil {
		return nil, nil, err
	}
	if file.UserID != userID {
		return nil, nil, fmt.Errorf("forbidden")
	}

	objectKey := file.Path
	if objectKey == "" {
		objectKey = file.Name
	}

	object, err := repository.GetObject(minioClient, objectKey)
	if err != nil {
		return nil, nil, err
	}

	return file, object, nil
}

func PreviewFile(ctx *gin.Context, minioClient *minio.Client) {
	fileID, err := strconv.Atoi(ctx.Param("fileID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID файла"})
		return
	}

	file, object, err := ownedFileObject(ctx, minioClient, fileID)
	if err != nil {
		if err.Error() == "unauthorized" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
			return
		}
		if err.Error() == "forbidden" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
			return
		}
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Файл не найден"})
		return
	}
	defer object.Close()

	kind := PreviewKind(file.Extension)
	if kind == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Предпросмотр недоступен для этого типа"})
		return
	}

	contentType := previewContentType(file.Name)
	ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", file.Name))
	ctx.Header("Content-Type", contentType)

	var reader io.Reader = object
	if isTextPreview(file.Extension) {
		reader = limitedTextReader(object, maxTextPreviewBytes)
	}
	if file.Size > 0 && kind != "text" {
		ctx.Header("Content-Length", fmt.Sprintf("%d", file.Size))
	}

	_, _ = io.Copy(ctx.Writer, reader)
}

func RenameFile(ctx *gin.Context) {
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

	newName := ctx.PostForm("newName")
	if newName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Имя не может быть пустым"})
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

	if err := repository.RenameFile(fileID, newName); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное имя файла"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Файл переименован"})
}
