package config

import "os"

var (
	MinioEndpoint  = envOr("MINIO_ENDPOINT", "localhost:9000")
	MinioAccessKey = envOr("MINIO_ACCESS_KEY", "minioadmin")
	MinioSecretKey = envOr("MINIO_SECRET_KEY", "minioadminpassword")
	MinioBucket    = envOr("MINIO_BUCKET", "user-files")
	MinioUseSSL    = false

	DBConnString = envOr("DATABASE_URL", "postgresql://postgres:6852@localhost:5432/cloud_storage")
	AppPort      = envOr("APP_PORT", "9091")
	JWTSecret    = envOr("JWT_SECRET", "your-very-secure-secret-key")
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
