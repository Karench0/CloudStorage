package repository

import (
	"CloudStorage/models"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var Ctx = context.Background()

func AddUser(form models.UserPass) {
	conn, _ := pgx.Connect(Ctx, "postgresql://postgres:6852@localhost:5432/cloud_storage")
	if err := conn.Ping(Ctx); err != nil {
		fmt.Println("не удалось подключиться к бд")
	} else {
		fmt.Println("бд успешно подключена!")
	}
	SqlExec := `
	INSERT INTO users (username, password)
	VALUES ($1, $2)
	`
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Exec(Ctx, SqlExec, form.Username, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func LoginUser(form models.UserPass) bool {
	var tmp models.UserPass
	conn, _ := pgx.Connect(Ctx, "postgresql://postgres:6852@localhost:5432/cloud_storage")
	SqlQuery := `
	SELECT * FROM users 
	WHERE username=$1
	`
	rows, _ := conn.Query(Ctx, SqlQuery, form.Username)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&tmp.Username, &tmp.Password)
	}
	fmt.Println("Для входа получены следующие значения", tmp)
	if CheckUser(form) == true && bcrypt.CompareHashAndPassword([]byte(tmp.Password), []byte(form.Password)) == nil {
		return true
	} else {
		return false
	}
}

func CheckUser(form models.UserPass) bool {
	var tmp models.UserPass
	conn, _ := pgx.Connect(Ctx, "postgresql://postgres:6852@localhost:5432/cloud_storage")
	SqlQuery := `
	SELECT username FROM users 
	WHERE username=$1
	`
	rows, _ := conn.Query(Ctx, SqlQuery, form.Username)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&tmp.Username)
	}
	if tmp.Username != "" {
		return true
	}
	return false
}
