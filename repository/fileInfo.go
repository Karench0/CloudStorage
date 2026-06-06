package repository

import (
	"CloudStorage/models"
	"fmt"
	"path/filepath"
	"time"
)

func SaveFile(file models.File) error {
	sqlQuery := `INSERT INTO files (name, size, path, user_id, directory_id, created_at) 
	             VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := DB.Exec(Ctx, sqlQuery, file.Name, file.Size, file.Path, file.UserID, file.DirectoryID, time.Now())
	if err != nil {
		fmt.Println("Ошибка при сохранении файла:", err)
		return err
	}
	return nil
}

func CreateDirectory(name string, userID int, parentID *int) error {
	sqlQuery := `INSERT INTO directories (name, user_id, parent_id, created_at) 
	             VALUES ($1, $2, $3, $4)`
	_, err := DB.Exec(Ctx, sqlQuery, name, userID, parentID, time.Now())
	if err != nil {
		fmt.Println("Ошибка при создании папки:", err)
		return err
	}
	return nil
}

func GetUserFiles(userID int) ([]models.File, error) {
	sqlQuery := `SELECT id, name, size, path, user_id, directory_id, created_at 
	             FROM files WHERE user_id=$1 AND directory_id IS NULL ORDER BY created_at DESC`
	rows, err := DB.Query(Ctx, sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(&file.ID, &file.Name, &file.Size, &file.Path, &file.UserID, &file.DirectoryID, &file.CreatedAt)
		if err != nil {
			continue
		}
		file.Extension = filepath.Ext(file.Name)
		files = append(files, file)
	}

	return files, nil
}

func GetUserDirectories(userID int) ([]models.Directory, error) {
	sqlQuery := `SELECT id, name, user_id, parent_id, size, created_at 
	             FROM directories WHERE user_id=$1 AND parent_id IS NULL ORDER BY created_at DESC`
	rows, err := DB.Query(Ctx, sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var directories []models.Directory
	for rows.Next() {
		var dir models.Directory
		err := rows.Scan(&dir.ID, &dir.Name, &dir.UserID, &dir.ParentID, &dir.Size, &dir.CreatedAt)
		if err != nil {
			continue
		}
		directories = append(directories, dir)
	}
	return directories, nil
}

func GetDirectoryByID(dirID int) (*models.Directory, error) {
	var dir models.Directory
	sqlQuery := `SELECT id, name, user_id, parent_id, size, created_at FROM directories WHERE id=$1`
	err := DB.QueryRow(Ctx, sqlQuery, dirID).Scan(&dir.ID, &dir.Name, &dir.UserID, &dir.ParentID, &dir.Size, &dir.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &dir, nil
}

func GetFileByID(fileID int) (*models.File, error) {
	var file models.File
	sqlQuery := `SELECT id, name, size, path, user_id, directory_id, created_at FROM files WHERE id=$1`
	err := DB.QueryRow(Ctx, sqlQuery, fileID).Scan(&file.ID, &file.Name, &file.Size, &file.Path, &file.UserID, &file.DirectoryID, &file.CreatedAt)
	if err != nil {
		return nil, err
	}
	file.Extension = filepath.Ext(file.Name)
	return &file, nil
}

func DeleteFileFromDB(fileID int) error {
	sqlQuery := `DELETE FROM files WHERE id=$1`
	_, err := DB.Exec(Ctx, sqlQuery, fileID)
	return err
}

func DeleteDirectory(dirID int) error {
	sqlQuery := `DELETE FROM directories WHERE id=$1`
	_, err := DB.Exec(Ctx, sqlQuery, dirID)
	return err
}

func RenameFile(fileID int, newName string) error {
	newName = filepath.Base(newName)
	if newName == "" || newName == "." {
		return fmt.Errorf("некорректное имя файла")
	}
	sqlQuery := `UPDATE files SET name=$1 WHERE id=$2`
	_, err := DB.Exec(Ctx, sqlQuery, newName, fileID)
	return err
}

type StorageStats struct {
	FileCount   int
	FolderCount int
	TotalBytes  int64
}

func GetUserStorageStats(userID int) (StorageStats, error) {
	var stats StorageStats
	err := DB.QueryRow(Ctx, `SELECT COUNT(*), COALESCE(SUM(size), 0) FROM files WHERE user_id=$1`, userID).
		Scan(&stats.FileCount, &stats.TotalBytes)
	if err != nil {
		return StorageStats{}, err
	}

	err = DB.QueryRow(Ctx, `SELECT COUNT(*) FROM directories WHERE user_id=$1`, userID).Scan(&stats.FolderCount)
	if err != nil {
		return StorageStats{}, err
	}

	return stats, nil
}

func RenameDirectory(dirID int, newName string) error {
	sqlQuery := `UPDATE directories SET name=$1 WHERE id=$2`
	_, err := DB.Exec(Ctx, sqlQuery, newName, dirID)
	return err
}

func GetFilesInDirectory(userID, dirID int) ([]models.File, error) {
	sqlQuery := `SELECT id, name, size, path, user_id, directory_id, created_at 
	             FROM files WHERE user_id=$1 AND directory_id=$2 ORDER BY created_at DESC`
	rows, err := DB.Query(Ctx, sqlQuery, userID, dirID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(&file.ID, &file.Name, &file.Size, &file.Path, &file.UserID, &file.DirectoryID, &file.CreatedAt)
		if err != nil {
			continue
		}
		file.Extension = filepath.Ext(file.Name)
		files = append(files, file)
	}
	return files, nil
}

func GetDirectoriesInDirectory(userID, dirID int) ([]models.Directory, error) {
	sqlQuery := `SELECT id, name, user_id, parent_id, size, created_at 
	             FROM directories WHERE user_id=$1 AND parent_id=$2 ORDER BY created_at DESC`
	rows, err := DB.Query(Ctx, sqlQuery, userID, dirID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var directories []models.Directory
	for rows.Next() {
		var dir models.Directory
		err := rows.Scan(&dir.ID, &dir.Name, &dir.UserID, &dir.ParentID, &dir.Size, &dir.CreatedAt)
		if err != nil {
			continue
		}
		directories = append(directories, dir)
	}
	return directories, nil
}

func GetDirectoryBreadcrumbs(dirID int) ([]models.Directory, error) {
	var breadcrumbs []models.Directory
	currentID := dirID

	for {
		dir, err := GetDirectoryByID(currentID)
		if err != nil {
			return breadcrumbs, err
		}
		breadcrumbs = append([]models.Directory{*dir}, breadcrumbs...)
		if dir.ParentID == nil {
			break
		}
		currentID = *dir.ParentID
	}
	return breadcrumbs, nil
}
