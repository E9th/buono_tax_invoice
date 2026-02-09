package config

import "os"

// Config เก็บค่า Configuration ทั้งหมดของระบบ
type Config struct {
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
	ServerPort  string
	DatabaseURL string // Render provides DATABASE_URL
}

// LoadConfig โหลด Config จาก Environment Variables (หรือใช้ค่า Default)
func LoadConfig() *Config {
	port := getEnv("PORT", "") // Render uses PORT
	if port == "" {
		port = getEnv("SERVER_PORT", "8080")
	}

	return &Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "buono_admin"),
		DBPassword:  getEnv("DB_PASSWORD", "buono_secure_2026"),
		DBName:      getEnv("DB_NAME", "buono_tax_invoice"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		ServerPort:  port,
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}
}

// GetDSN สร้าง Connection String สำหรับ PostgreSQL
// ถ้ามี DATABASE_URL (Render) จะใช้ตัวนั้นก่อน
func (c *Config) GetDSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
