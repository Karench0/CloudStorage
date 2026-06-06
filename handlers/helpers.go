package handlers

import (
	"CloudStorage/models"
	"CloudStorage/repository"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func currentUser(ctx *gin.Context) (username string, userID int, ok bool) {
	u, existsUrl := ctx.Get("username")
	id, existsID := ctx.Get("userID")
	if !existsUrl || !existsID {
		return "", 0, false
	}
	return u.(string), id.(int), true
}

func renderBrowse(ctx *gin.Context, data gin.H) {
	if _, userID, ok := currentUser(ctx); ok {
		if stats, err := repository.GetUserStorageStats(userID); err == nil {
			data["Stats"] = stats
		}
	}

	ctx.HTML(http.StatusOK, "dashboard.html", data)
}

func redirectBrowse(ctx *gin.Context, directoryID *int) {
	if directoryID != nil {
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/folder/%d", *directoryID))
		return
	}
	ctx.Redirect(http.StatusSeeOther, "/dashboard")
}

func setUploadErrorAndRedirect(ctx *gin.Context, message string, directoryID *int) {
	if directoryID != nil {
		ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/folder/%d?error=%s", *directoryID, message))
		return
	}
	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/dashboard?error=%s", message))
}

func parseDirectoryID(ctx *gin.Context, userID int) *int {
	dirIDStr := ctx.Query("dirID")
	if dirIDStr == "" {
		dirIDStr = ctx.PostForm("dirID")
	}
	if dirIDStr == "" {
		return nil
	}

	var dirID int
	if _, err := fmt.Sscanf(dirIDStr, "%d", &dirID); err != nil {
		return nil
	}

	dir, err := repository.GetDirectoryByID(dirID)
	if err != nil || dir == nil || dir.UserID != userID {
		return nil
	}

	return &dirID
}

func loadRootBrowse(userID int) (files []models.File, dirs []models.Directory, err error) {
	files, err = repository.GetUserFiles(userID)
	if err != nil {
		files = make([]models.File, 0)
	}
	dirs, err = repository.GetUserDirectories(userID)
	if err != nil {
		dirs = make([]models.Directory, 0)
	}
	return files, dirs, err
}

func loadFolderBrowse(userID, folderID int) (folder *models.Directory, files []models.File, dirs []models.Directory, breadcrumbs []models.Directory, err error) {
	folder, err = repository.GetDirectoryByID(folderID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	files, err = repository.GetFilesInDirectory(userID, folderID)
	if err != nil {
		files = make([]models.File, 0)
	}

	dirs, err = repository.GetDirectoriesInDirectory(userID, folderID)
	if err != nil {
		dirs = make([]models.Directory, 0)
	}

	breadcrumbs, err = repository.GetDirectoryBreadcrumbs(folderID)
	if err != nil {
		breadcrumbs = []models.Directory{*folder}
	}

	return folder, files, dirs, breadcrumbs, nil
}
