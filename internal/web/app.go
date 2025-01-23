package web

import (
	"fmt"
	"net/http"
	"support-bot/internal/cache"
	"support-bot/internal/log"
	"support-bot/internal/request"
	googlesheet "support-bot/internal/sheet"
	"support-bot/internal/tg"
	"support-bot/internal/usedesk"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

// WebApp веб приложение. Мозг программы, который использует большинство других приложений
type WebApp struct {
	config Config // Параметры

	Router        *mux.Router                   // Маршрутизатор
	TemplateCache map[string]*template.Template // Карта шаблонов

	log.Logger

	usedesk     *usedesk.ClientUsedesk
	ticketCache *cache.TicketCache

	Bot   *tg.Bot                  // Интерфейс бота
	Sheet googlesheet.SheetManager // Интерфейс таблиц
	Cache cache.Cacher             // Интерфейс кеша
}

// NewWebApp создает и возвращает веб приложение
func NewWebApp(config Config, bot *tg.Bot, sheet googlesheet.SheetManager, ticketCache *cache.TicketCache, usedesk *usedesk.ClientUsedesk) (*WebApp, error) {
	// Загрузка шаблонов
	templateCache, err := NewTemplateCache("./ui/html/")
	if err != nil {
		return nil, err
	}

	// Запуск обработчиков юздеск
	reqUsedesk, err := request.NewRequestHandler(log.Log, 100)
	if err != nil {
		return nil, err
	}
	go reqUsedesk.ProcessRequests(500 * time.Millisecond)

	app := WebApp{
		Logger:        log.Log,
		config:        config,
		Bot:           bot,
		Sheet:         sheet,
		Cache:         cache.NewCachedData(),
		TemplateCache: templateCache,
		usedesk:       usedesk,
		ticketCache:   ticketCache,
	}

	err = app.UpdateCache()
	if err != nil {
		return nil, err
	}
	go app.StartPeriodUpdateCache(24 * time.Hour)

	// Установка параметров
	app.Router = app.SetRoutes()
	return &app, nil
}

// HandleUpdates запускает HTTP сервер
func (app *WebApp) StartServer() error {
	app.Info("Запуск сервера по адрессу " + app.config.IP + ":" + app.config.PORT)
	err := http.ListenAndServe(app.config.IP+":"+app.config.PORT, app.Router)
	if err != nil {
		return fmt.Errorf("ошибка при запуске сервера: %v", err)
	}
	return nil
}

func (app *WebApp) ValidateInitData(initDataStr, token string) (*initdata.InitData, error) {
	expIn := 1 * time.Hour
	err := initdata.Validate(initDataStr, token, expIn)
	if err != nil {
		return nil, err
	}
	initData, err := initdata.Parse(initDataStr)
	if err != nil {
		return nil, err
	}
	return &initData, nil
}
