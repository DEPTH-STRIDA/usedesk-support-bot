package tg

import (
	"errors"
	"fmt"
	"support-bot/internal/cache"
	"support-bot/internal/log"
	"support-bot/internal/request"
	"support-bot/internal/usedesk"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Request func() error

type Bot struct {
	log.Logger
	*tgbotapi.BotAPI
	config                 Config
	msgRequestHandler      *request.RequestHandler
	callbackRequestHandler *request.RequestHandler
	MsgRoutes              map[string]HandlerFunc
	CallbackRoutes         map[string]HandlerFunc

	ticketCache *cache.TicketCache
	usedesk     *usedesk.ClientUsedesk
}

// Конструктор нового бота
func NewBot(config Config, ticketCache *cache.TicketCache, usedesk *usedesk.ClientUsedesk) (*Bot, error) {
	msgRequestHandler, err := request.NewRequestHandler(log.Log, int64(config.MsgBufferSize))
	if err != nil {
		return nil, err
	}
	callbackRequestHandler, err := request.NewRequestHandler(log.Log, int64(config.CallBackBufferSize))
	if err != nil {
		return nil, err
	}
	app := Bot{
		config:                 config,
		Logger:                 log.Log,
		msgRequestHandler:      msgRequestHandler,
		callbackRequestHandler: callbackRequestHandler,
		ticketCache:            ticketCache,
		usedesk:                usedesk,
	}

	go app.msgRequestHandler.ProcessRequests(time.Duration(app.config.RequestUpdatePause) * time.Second)
	go app.callbackRequestHandler.ProcessRequests(time.Duration(app.config.RequestCallBackUpdatePause) * time.Second)

	app.BotAPI, err = tgbotapi.NewBotAPI(app.config.Token)
	if err != nil {
		return nil, fmt.Errorf("не удается инициализировать бота telegram: %v", err)
	}

	go app.HandleUpdates()

	app.MsgRoutes = make(map[string]HandlerFunc)
	app.CallbackRoutes = make(map[string]HandlerFunc)

	err = app.SetMsgRoutes()
	if err != nil {
		return nil, errors.New("установка обработчиков сообщений не удалась: " + err.Error())
	}
	err = app.SetCallBackRoutes()
	if err != nil {
		return nil, errors.New("установка обработчиков callback событий не удалась: " + err.Error())
	}
	return &app, nil
}

// HandleUpdates запускает обработку всех обновлений поступающих боту из телеграмма
func (app *Bot) HandleUpdates() {
	// Настройка обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	// Получение канала обновлений
	updates := app.GetUpdatesChan(u)
	for update := range updates {

		if update.Message != nil {
			switch {
			case update.Message.MediaGroupID != "":
				go app.handleMediaGroup(update)
			case update.Message != nil:
				go app.handleMessage(update)
			default:
				app.Info("Боту поступило неизвестное обновление")
			}
		}

		if update.CallbackQuery != nil {
			go app.handleCallback(update)
		}
	}
}

// handleMessage ищет команду в map'е и выполняет ее
func (app *Bot) handleMessage(update tgbotapi.Update) {
	app.Info("Обработка сообщения от: ", update.Message.Chat.ID)
	if update.Message.Chat.ID == app.config.BotTgChat || update.Message.Chat.ID == app.config.SupportChatID {
		app.Error("Игнорирование обновления из главных чатов: ", update.Message.Chat.ID)
		return
	}
	currentAction, ok := app.MsgRoutes[update.Message.Text]
	if ok {
		err := currentAction(app, update)
		if err != nil {
			app.Error("Ошибка при обработки команды ", update.Message.Text, " от пользователя (", update.Message.Chat.ID, ":", update.Message.Chat.UserName)
		} else {
			app.Info("Успешно обработана команда: ", update.Message.Text, " от пользователя (", update.Message.Chat.ID, ":", update.Message.Chat.UserName)

		}
		return
	}

	// обработка любых сообщения от пользователя, если они содержат только текст
	if update.Message.Text != "" {
		app.HandleUserMessage(update)
	}
}

// handleCallback ищет команду в map'е и выполняет ее
func (app *Bot) handleCallback(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}
	// Вызов стандартных функций из БД
	currentAction, ok := app.CallbackRoutes[update.CallbackQuery.Data]
	if !ok {
		//  Запуск события на обработку реакции
		currentActionTemp, ok := app.CallbackRoutes["callbackDB"]
		if !ok {
			// Если метода для обработки реакции нет, значит явно есть ошибка
			app.Info("Не удалось запустить обработку callbackDB! Проверьте установку метода в пакете notification")
			app.Info("Неизвестная Callback команда")
			return
		} else {
			currentAction = currentActionTemp
		}
	}
	err := currentAction(app, update)
	if err != nil {
		app.Error("Ошибка при обработки Callback команды от пользователя (", update.CallbackQuery.From.ID, ":", update.CallbackQuery.From.UserName)
	} else {
		app.Info("Успешно обработана Callback команда: ", update.CallbackQuery.Data, " от пользователя (", update.CallbackQuery.From.ID, ":", update.CallbackQuery.From.UserName)
	}
}

// HandlRoute добавляет обработку комманды
func (app *Bot) HandlMsgRoute(command string, handler HandlerFunc) {
	app.MsgRoutes[command] = handler
}

// HandlRoute добавляет обработку комманды
func (app *Bot) HandlCallbackRoute(command string, handler HandlerFunc) {
	app.CallbackRoutes[command] = handler
	app.Info("Успешно установлена callback команда: ", command)
}

func (app *Bot) DeleteCallbackRoute(command string) {
	delete(app.CallbackRoutes, "command")
}

func (app *Bot) DeleteMsgRoute(command string) {
	delete(app.MsgRoutes, "command")
}

func (app *Bot) GetConfig() Config {
	return app.config
}
