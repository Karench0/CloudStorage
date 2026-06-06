package repository

import (
	"CloudStorage/config"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	DB  *pgxpool.Pool
	Ctx = context.Background()
)

func InitDB() {
	var err error
	DB, err = pgxpool.New(Ctx, config.DBConnString)
	if err != nil {
		log.Fatalf("Не удалось создать пул подключений к БД: %v", err)
	}

	if err = DB.Ping(Ctx); err != nil {
		log.Fatalf("Нет ответа от БД при пинге пула: %v", err)
	}
	log.Println("Пул подключений к БД успешно инициализирован!")
}
