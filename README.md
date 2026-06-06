# Облачное хранилище

Веб-приложение на Go (Gin) + PostgreSQL + MinIO.

## Запуск инфраструктуры

Из корня проекта:

```bash
docker compose up -d
```

Поднимаются:

| Сервис   | Порт  | Назначение |
|----------|-------|------------|
| Postgres | 5432  | метаданные (пользователи, папки, файлы) |
| MinIO API | **9000** | хранение файлов (использует приложение) |
| MinIO Console | **9001** | веб-интерфейс в браузере |
| App | **9091** | Само приложение |


Веб-консоль MinIO:

- URL: **http://localhost:9001**
- Логин: `minioadmin`
- Пароль: `minioadminpassword`

(те же значения, что в `docker-compose.yaml` и `config/config.go`.)

Проверка, что MinIO запущен:

```bash
docker compose ps
curl -s http://127.0.0.1:9000/minio/health/live
```

Должен вернуть пустой ответ с кодом `200`.

Если контейнер не запущен:

```bash
docker compose up -d minio
docker logs cloud_storage_minio
```

Приложение: **http://localhost:9091**

Перед запуском должны работать Postgres и MinIO.

## Структура проекта

```
main.go              — точка входа
config/              — настройки (env или значения по умолчанию)
handlers/            — HTTP: страницы, загрузка, API удаления
  helpers.go         — сессия, отрисовка каталога, ошибки загрузки
  files.go           — upload / download
  dashboard.go       — папки и список файлов
  routes.go          — регистрация маршрутов
repository/          — БД и MinIO (без HTTP)
models/              — типы данных
templates/           — HTML
docker-compose.yaml  — Postgres + MinIO
```
