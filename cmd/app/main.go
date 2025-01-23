package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"support-bot/internal/cache"
	"support-bot/internal/log"
	"support-bot/internal/model"
	"support-bot/internal/usedesk"

	googlesheet "support-bot/internal/sheet"
	"support-bot/internal/tg"
	"support-bot/internal/web"
)

type Config struct {
	Web      web.Config         `json:"webb-app"`
	Telegram tg.Config          `json:"telegram"`
	Sheet    googlesheet.Config `json:"google-sheet"`
}

func main() {
	if err := run(); err != nil {
		fmt.Println("Критическая остановка приложения: ", err)
		os.Exit(1)
	}
}

func run() error {
	// Установка МСК часового пояса для приложения.
	err := setLocationTime("Europe/Moscow")
	if err != nil {
		return err
	}

	// Создание логгера. Логи будут писаться в папке log.
	log.Log, err = log.NewCombinedLogger("log/")
	if err != nil {
		return fmt.Errorf("ошибка при создании логгера: %w", err)
	}

	// Анмаршалинг конфигурационного файла.
	var config Config
	if err := readJSONFile("config/config.json", &config); err != nil {
		return fmt.Errorf("ошибка при анмаршалинге конфигурационного файла: %w", err)
	}

	v, _ := json.Marshal(config)
	fmt.Println(string(v))

	model.DataBaseManager, err = model.NewDBManager()
	if err != nil {
		return fmt.Errorf("ошибка приподключении к БД: %w", err)
	}

	// Установка порта и адреса. Если в конфиге нет порта и адреса, то они будут браться из окружения.
	if config.Web.IP == "" {
		exist := false
		config.Web.IP, exist = os.LookupEnv("APP_IP")
		if !exist || config.Web.IP == "" {
			return fmt.Errorf("не удалось получить переменную среды APP_IP или переменная пуста")
		}
	}
	if config.Web.PORT == "" {
		exist := false
		config.Web.PORT, exist = os.LookupEnv("APP_PORT")
		if !exist {
			return fmt.Errorf("не удалось получить переменную среды APP_PORT или переменная пуста")
		}
	}

	ticketCache := cache.NewTicketCache(model.DataBaseManager.DB)
	if ticketCache == nil {
		return fmt.Errorf("не удалось создать кеш")
	}

	usedesk, err := usedesk.NewClient(config.Web.TokenUsdesk)
	if err != nil {
		return err
	}

	// Создание и запуск телеграм бота
	bot, err := tg.NewBot(config.Telegram, ticketCache, usedesk)
	if err != nil {
		return fmt.Errorf("создание телеграмм бота не удалось: %w", err)
	}

	// Создание Google Sheets сервиса
	gSheet, err := googlesheet.NewGoogleSheets(config.Sheet)
	if err != nil {
		return fmt.Errorf("создание сервиса Google Sheets не удалось: %w", err)
	}

	// Создание веб приложения
	webApp, err := web.NewWebApp(config.Web, bot, gSheet, ticketCache, usedesk)
	if err != nil {
		return fmt.Errorf("создание веб приложения не удалось: %w", err)
	}
	log.Log.Info("Веб приложение успешно запущено")

	googlesheet.SetUserData(gSheet)

	go func() {
		for {
			time.Sleep(12 * time.Hour)
			googlesheet.SetUserData(gSheet)
		}

	}()

	// Обработка обновлений
	webApp.StartServer()
	return nil
}

// setLocationTime устанавливает часовой пояс по умолчанию для глобальной переменной.
// Принимает строку с названием локации и возвращает ошибку, если часовой пояс не удалось загрузить.
func setLocationTime(location string) error {
	// Устанавливаем локацию по умолчанию для time.Local
	loc, err := time.LoadLocation(location)
	if err != nil {
		return fmt.Errorf("ошибка при смене локации на %s: %w", location, err)
	}
	time.Local = loc
	return nil
}

// readJSONFile принимает имя файла и указатель на структуру, в которую будет распарсен JSON
func readJSONFile(filename string, v interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	if err := json.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("ошибка анмаршалинга JSON: %w", err)
	}

	return nil
}
