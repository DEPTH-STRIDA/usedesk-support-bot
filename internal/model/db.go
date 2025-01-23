package model

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Глобальная переменная для взаимодействия с другими пакетами
var DataBaseManager *DBManager

// Менеджер БД
type DBManager struct {
	*gorm.DB
}

// NewDBManager выполняет подключение к БД по данным пользователя из окружения.
func NewDBManager() (*DBManager, error) {
	godotenv.Load()

	var dbUri string
	connString := os.Getenv("DB_CONNECTION_STRING")

	if connString != "" {
		dbUri = connString
	} else {
		username := os.Getenv("DBUSER")
		if username == "" {
			return nil, fmt.Errorf("db_user is empty")
		}

		password := os.Getenv("DBPASS")
		if password == "" {
			return nil, fmt.Errorf("db_pass is empty")
		}

		dbName := os.Getenv("DBNAME")
		if dbName == "" {
			return nil, fmt.Errorf("db_name is empty")
		}

		dbHost := os.Getenv("DBHOST")
		if dbHost == "" {
			return nil, fmt.Errorf("db_host is empty")
		}

		dbPort := os.Getenv("DBPORT")
		if dbPort == "" {
			dbPort = "5432" // Порт по умолчанию для PostgreSQL
		}

		// Формат строки подключения key=value
		dbUri = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			dbHost, username, password, dbName, dbPort)
	}

	fmt.Println("Строка подключения к БД: ", dbUri)

	// Подключение к БД
	conn, err := gorm.Open(postgres.Open(dbUri), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Создаем менеджер БД
	dbManager := &DBManager{DB: conn}

	// Миграция структур в таблицы в БД, если их там не было.
	conn.Debug().AutoMigrate(&Ticket{})

	return dbManager, nil
}
