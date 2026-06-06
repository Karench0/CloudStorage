package models

import "time"

type UserPass struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

type File struct {
	ID          int
	Name        string
	Size        int64
	Path        string
	UserID      int
	DirectoryID *int
	CreatedAt   time.Time
	Extension   string
	ContentType string
}

type Directory struct {
	ID        int
	Name      string
	UserID    int
	ParentID  *int
	Size      int64
	CreatedAt time.Time
}
