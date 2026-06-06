package repository

import (
	"CloudStorage/models"

	"golang.org/x/crypto/bcrypt"
)

func GetUserID(username string) (int, error) {
	var userID int
	err := DB.QueryRow(Ctx, "SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func AddUser(form models.UserPass) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	SqlExec := `INSERT INTO users (username, password_hash) VALUES ($1, $2)`
	_, err = DB.Exec(Ctx, SqlExec, form.Username, hashedPassword)
	return err
}

func LoginUser(form models.UserPass) (int, bool) {
	var userID int
	var dbPasswordHash string

	SqlQuery := `SELECT id, password_hash FROM users WHERE username=$1`
	err := DB.QueryRow(Ctx, SqlQuery, form.Username).Scan(&userID, &dbPasswordHash)
	if err != nil {
		return 0, false
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbPasswordHash), []byte(form.Password))
	if err != nil {
		return 0, false
	}
	return userID, true
}

func CheckUser(form models.UserPass) bool {
	var username string
	SqlQuery := `SELECT username FROM users WHERE username=$1`
	err := DB.QueryRow(Ctx, SqlQuery, form.Username).Scan(&username)

	return err == nil && username != ""
}
