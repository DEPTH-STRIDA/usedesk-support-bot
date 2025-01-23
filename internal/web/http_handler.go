package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	googlesheet "support-bot/internal/sheet"
	"support-bot/internal/usedesk"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/microcosm-cc/bluemonday"
)

func (app *WebApp) DebugHandler(w http.ResponseWriter, r *http.Request) {
	PrintRequest(r)
	app.Info("Starting to process incoming request")

	event, err := usedesk.DetermineEventType(r)
	if err != nil {
		app.Error("Error determining event type: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	app.Info("Determined event type: %s", event.Type)

	switch event.Type {
	case "message":
		// Обработка сообщения от агента через форму
		messageEvent := event.Payload.(usedesk.MessageEvent)
		app.Info("Received message event: %+v", messageEvent)

		// Поскольку в форме нет ticket_id, используем следующий JSON
		if messageEvent.TicketID == 0 {
			app.Info("Skipping form message, waiting for JSON with ticket_id")
			break
		}

		telegramID, ok := app.ticketCache.GetTelegramByAnyTicket(messageEvent.TicketID)
		if !ok {
			app.Error("Failed to get Telegram ID for ticket ID: %d", messageEvent.TicketID)
			break
		}

		app.Info("Attempting to send message to Telegram ID: %d", telegramID)
		cleanMessage := stripHTMLTags(messageEvent.Message)
		msg := tgbotapi.NewMessage(telegramID, cleanMessage)
		msg.ParseMode = "html"

		if _, err := app.Bot.Send(msg); err != nil {
			app.Error("Error sending message to Telegram: %v", err)
		} else {
			app.Info("Successfully sent message to Telegram ID: %d", telegramID)
		}

	case "comment":
		app.Info("Processing comment event")
		commentEvent := event.Payload.(usedesk.CommentEvent)
		app.Info("Comment details - TicketID: %d, From: %s, Type: %s, Message: %s",
			commentEvent.Comment.TicketID,
			commentEvent.Comment.From,
			commentEvent.Comment.Type,
			commentEvent.Comment.Message)

		// Игнорируем комментарии от клиента и приватные комментарии
		if commentEvent.Comment.From == "client" {
			app.Info("Ignoring client comment for ticket %d", commentEvent.Comment.TicketID)
			break
		}
		if commentEvent.Comment.Type == "private" {
			app.Info("Ignoring private comment for ticket %d", commentEvent.Comment.TicketID)
			break
		}

		app.Info("Comment is public and from agent, proceeding with processing")

		// Получаем Telegram ID пользователя по Ticket ID
		app.Info("Attempting to get Telegram ID for ticket %d", commentEvent.Comment.TicketID)
		telegramID, ok := app.ticketCache.GetTelegramByAnyTicket(commentEvent.Comment.TicketID)
		if !ok {
			app.Error("Failed to get Telegram ID for ticket ID: %d", commentEvent.Comment.TicketID)
			break
		}
		app.Info("Successfully got Telegram ID: %d for ticket: %d", telegramID, commentEvent.Comment.TicketID)

		// Очистка сообщения
		app.Info("Cleaning HTML from message")
		cleanMessage := stripHTMLTags(commentEvent.Comment.Message)
		app.Info("Original message: %s", commentEvent.Comment.Message)
		app.Info("Cleaned message: %s", cleanMessage)

		// Подготовка и отправка сообщения в Telegram
		app.Info("Preparing Telegram message for user %d", telegramID)
		msg := tgbotapi.NewMessage(telegramID, cleanMessage)
		msg.ParseMode = "html"

		app.Info("Attempting to send Telegram message to user %d", telegramID)
		if _, err := app.Bot.Send(msg); err != nil {
			app.Error("Failed to send Telegram message: %v", err)
			app.Error("Message details - UserID: %d, Message: %s", telegramID, cleanMessage)
		} else {
			app.Info("Successfully sent Telegram message to user %d", telegramID)
			app.Info("Message content: %s", cleanMessage)
		}

	case "trigger":
		app.Info("Processing trigger event")
		triggerEvent := event.Payload.(usedesk.TriggerEvent)
		app.Info("Trigger details - TicketID: %d, NewStatus: %s",
			triggerEvent.Trigger.TicketID,
			triggerEvent.Trigger.NewStatus)

		app.handleTriggerEvent(triggerEvent)

	default:
		app.Info("Received unsupported event type: %s", event.Type)
	}

	app.Info("Finished processing request")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Event processed")
}

func (app *WebApp) handleTriggerEvent(triggerEvent usedesk.TriggerEvent) {
	app.Info("Processing trigger event: %+v", triggerEvent)
	newStatus := triggerEvent.Trigger.NewStatus
	app.Info("Trigger event STATUS: %v", newStatus)

	// Получаем Telegram ID пользователя по Ticket ID
	userID, ok := app.ticketCache.GetTelegramByAnyTicket(triggerEvent.Trigger.TicketID)
	if !ok {
		app.Error("Failed to get Telegram ID for ticket ID: %d", triggerEvent.Trigger.TicketID)
		return
	}

	app.Info("Got userID: %d for ticketID: %d", userID, triggerEvent.Trigger.TicketID)
	app.Info("Processing trigger event for userID: %d with new status: %s", userID, newStatus)

	if newStatus == "0" || newStatus == "1" || newStatus == "5" || newStatus == "6" || newStatus == "8" {
		app.Info("No action required for status: %s", newStatus)
	} else {
		app.Info("Deleting ticket %d due to status change to %s", triggerEvent.Trigger.TicketID, newStatus)
		if err := app.ticketCache.DeleteTicket(triggerEvent.Trigger.TicketID); err != nil {
			app.Error("Failed to delete ticket: %v", err)
		} else {
			app.Info("Successfully deleted ticket %d", triggerEvent.Trigger.TicketID)
		}
	}
}

func stripHTMLTags(input string) string {
	p := bluemonday.StripTagsPolicy()
	cleaned := p.Sanitize(input)
	return strings.TrimSpace(cleaned)
}

func PrintRequest(r *http.Request) {
	fmt.Println("--- Begin Request ---")
	fmt.Printf("Method: %s\n", r.Method)
	fmt.Printf("URL: %s\n", r.URL.String())

	fmt.Println("\n--- Headers ---")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}

	fmt.Println("\n--- Query Parameters ---")
	for name, values := range r.URL.Query() {
		for _, value := range values {
			fmt.Printf("%s: %s\n", name, value)
		}
	}

	fmt.Println("\n--- Body ---")
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("Error reading body: %v\n", err)
		} else {
			bodyString := string(bodyBytes)
			fmt.Println(bodyString)

			// Восстанавливаем тело запроса
			r.Body = io.NopCloser(strings.NewReader(bodyString))

			// Попытка разобрать JSON
			if r.Header.Get("Content-Type") == "application/json" {
				var jsonData interface{}
				if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
					fmt.Println("\n--- Parsed JSON ---")
					jsonPretty, _ := json.MarshalIndent(jsonData, "", "  ")
					fmt.Println(string(jsonPretty))
				}
			}

			// Обработка form-data остается без изменений
		}
	} else {
		fmt.Println("No body")
	}

	fmt.Println("--- End Request ---")
}

func (app *WebApp) handleMessageEvent(messageEvent usedesk.MessageEvent) {
	app.Info("Received message event: %+v", messageEvent)

	// Получаем telegram ID по ticket ID через существующий кэш
	telegramID, ok := app.ticketCache.GetTelegramByAnyTicket(messageEvent.TicketID)
	if !ok {
		app.Error("Failed to get Telegram ID for ticket ID: %d", messageEvent.TicketID)
		return
	}

	app.Info("Attempting to send message to Telegram ID: %d", telegramID)

	cleanMessage := stripHTMLTags(messageEvent.Message)
	msg := tgbotapi.NewMessage(telegramID, cleanMessage)
	msg.ParseMode = "html"

	if _, err := app.Bot.Send(msg); err != nil {
		app.Error("Error sending message to Telegram: %v", err)
	} else {
		app.Info("Message sent successfully to Telegram ID: %d", telegramID)
	}
}

// Добавляем новый обработчик для комментариев
func (app *WebApp) handleCommentEvent(commentEvent usedesk.CommentEvent) {
	app.Info("Received comment event: %+v", commentEvent) // Логируем полученное событие комментария

	// Игнорируем комментарии от клиента
	if commentEvent.Comment.From == "client" {
		app.Info("Игнорирование создания сообщения от клиента") // Логируем игнорирование
		return
	}

	// Получаем Telegram ID пользователя по Ticket ID
	userID, ok := app.ticketCache.GetTelegramByAnyTicket(commentEvent.Comment.TicketID)
	if !ok {
		app.Error("Failed to get Telegram ID for ticket ID: %d", commentEvent.Comment.TicketID)
		return
	}

	app.Info("Attempting to send message to userID: %d", userID)

	// Проверяем, что userID не равен 0
	if userID != 0 {
		cleanMessage := stripHTMLTags(commentEvent.Comment.Message)
		msg := tgbotapi.NewMessage(userID, cleanMessage)
		msg.ParseMode = "html"

		// Отправляем сообщение в Telegram
		if _, err := app.Bot.Send(msg); err != nil {
			app.Error("Error sending message to Telegram: %v", err)
		} else {
			app.Info("Message sent successfully to userID: %d", userID)
		}
	} else {
		app.Warn("Comment event received with UserID = 0")
	}
}

var version = fmt.Sprintf("v%d", time.Now().Unix())

// HandleValidate главная страница, которая пересылает пользователя на страницу с формой и отправляет url query
func (app *WebApp) HandleValidate(w http.ResponseWriter, r *http.Request) {
	// Генерируем версию или используем существующую
	version := fmt.Sprintf("%d", time.Now().Unix())

	data := map[string]interface{}{
		"Version": version,
	}

	// рендер шаблона с данными
	err := app.render(w, "validate.page.tmpl", data)
	if err != nil {
		app.Error("HandleValidate. Не удалось выполнить рендер: ", err)
		http.Error(w, "HandleValidate. Не удалось выполнить рендер: "+err.Error(), http.StatusInternalServerError)
	}
}

// HandleValidate главная страница, которая пересылает пользователя на страницу с формой и отправляет url query
func (app *WebApp) HandleForm(w http.ResponseWriter, r *http.Request) {
	initdata, err := app.ValidateInitData(r.URL.Query().Get("initData"), app.Bot.GetConfig().Token)
	if err != nil {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleAdminMenu. Неверные телеграмм данные:" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерируем версию или используем существующую

	cache := app.Cache.GetData()
	typedCache, ok := cache.(GoogleSheetData)
	if !ok {
		app.Warn("(" + r.RemoteAddr + ") Не удалось привести кеш к нормальному типу")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	name := googlesheet.GetUserName(initdata.User.Username)

	data := map[string]interface{}{
		"Version":  version,             //string
		"Tegs":     typedCache.Tegs,     //[]string
		"Problems": typedCache.Problems, //[][]string
		"Name":     name,
	}

	fmt.Println(data)

	// рендер шаблона с данными
	err = app.render(w, "form.page.tmpl", data)
	if err != nil {
		app.Error("HandleForm. Не удалось выполнить рендер: ", err)
		http.Error(w, "HandleForm. Не удалось выполнить рендер: "+err.Error(), http.StatusInternalServerError)
	}
}

func (app *WebApp) HandleIsAdmin(w http.ResponseWriter, r *http.Request) {
	initData, err := app.ValidateInitData(r.URL.Query().Get("initData"), app.Bot.GetConfig().Token)
	if err != nil {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleIsAdmin. Неверные телеграмм данные:" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !containsGeneric(app.Bot.GetConfig().Admins, initData.User.ID) {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleIsAdmin. Неверные телеграмм данные:" + err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Обобщенная функция для проверки наличия элемента в слайсе (Go 1.18+)
func containsGeneric[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// HandleAdminMenu
func (app *WebApp) HandleAdminMenu(w http.ResponseWriter, r *http.Request) {
	app.Info(r.URL.Query().Get("initData"))
	app.Info(r.URL)
	initData, err := app.ValidateInitData(r.URL.Query().Get("initData"), app.Bot.GetConfig().Token)
	if err != nil {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleAdminMenu. Неверные телеграмм данные:" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !containsGeneric(app.Bot.GetConfig().Admins, initData.User.ID) {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleAdminMenu. Неверные телеграмм данные:" + err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// Генерируем версию или используем существующую
	version := fmt.Sprintf("%d", time.Now().Unix())

	data := map[string]interface{}{
		"Version": version,
	}

	// рендер шаблона с данными
	err = app.render(w, "admin-menu.page.tmpl", data)
	if err != nil {
		app.Error("HandleAdminMenu. Не удалось выполнить рендер: ", err)
		http.Error(w, "HandleAdminMenu. Не удалось выполнить рендер: "+err.Error(), http.StatusInternalServerError)
	}
}

func (app *WebApp) HandleAdminCommand(w http.ResponseWriter, r *http.Request) {
	// app.Info(r.URL)
	initData, err := app.ValidateInitData(r.URL.Query().Get("initData"), app.Bot.GetConfig().Token)
	if err != nil {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleAdminMenu. Неверные телеграмм данные:" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !containsGeneric(app.Bot.GetConfig().Admins, initData.User.ID) {
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleAdminMenu. Неверные телеграмм данные:")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	commad := r.URL.Query().Get("command")

	switch commad {
	case "update-select-data":
		app.UpdateCache()
	default:
		app.Warn("(" + r.RemoteAddr + ") Неизвестная команда:" + commad)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Данные тегов и проблем успешно обновились")
}

// func (app *WebApp) preapareMsg(form Form) tgbotapi.MessageConfig {
// 	// Тег
// 	msgText := app.preapareMsgText(form)

// 	msg := tgbotapi.NewMessage(app.Bot.GetConfig().SupportChatID, msgText)
// 	return msg
// }

func (app *WebApp) preapareMsgText(form Form) string {
	// Тег
	msgText := "#" + form.Department + "\n"
	if form.IsEmergency {
		msgText += "#срочно" + "\n"
	}
	// ФИ препода
	msgText += form.Name + " @" + form.UserName + "\n"
	// Где проходит занятие
	if strings.Trim(form.Place, " ") != "" {
		msgText += form.Place + "\n"
	}
	// Номер группы
	if strings.Trim(form.GroupNumber, " ") != "" {
		msgText += fmt.Sprint(form.GroupNumber) + "\n"
	}
	// Проблема
	msgText += form.ReadyProblem + "\n"
	// Кастомная проблема
	if strings.Trim(form.CustomProblem, " ") != "" {
		msgText += form.CustomProblem + "\n"
	}
	return msgText
}

func (app *WebApp) HandleSendData(w http.ResponseWriter, r *http.Request) {
	// Читаем тело запроса
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		// Логируем ошибку, если не удалось прочитать тело
		app.Error("HandleSendData. Не удалось прочитать тело: ", err)
		http.Error(w, "HandleSendData. Не удалось прочитать тело: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Логируем полученные данные формы
	app.Info("Пришла форма для обработки: ", string(bodyBytes))

	var form Form
	// Разбираем JSON из тела запроса в структуру Form
	if err := json.Unmarshal(bodyBytes, &form); err != nil {
		// Логируем ошибку, если не удалось разобрать JSON
		app.Error("HandleSendData. Ошибка при разборе JSON: ", err)
		http.Error(w, "HandleSendData. Ошибка при разборе JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем и валидируем данные инициализации
	initData, err := app.ValidateInitData(form.InitData, app.Bot.GetConfig().Token)
	if err != nil {
		// Логируем предупреждение о неверных данных
		app.Warn("(" + r.RemoteAddr + ") Вход запрещен. HandleSendData. Неверные телеграмм данные:" + err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Сохраняем имя пользователя из инициализации в форму
	form.UserName = initData.User.Username
	// Подготавливаем текст сообщения для тикета
	msg := app.preapareMsgText(form)

	// Логируем создание нового тикета
	app.Info("Создание нового тикета для пользователя: ", initData.User.Username)

	// Создаем новый запрос на создание тикета
	newTicket := usedesk.NewTicketRequest{
		Subject:    form.Department,
		Message:    msg,
		ClientName: initData.User.FirstName + " " + initData.User.LastName + "(" + initData.User.Username + ")",
		Type:       "question",
	}

	// Отправляем запрос на создание тикета
	response, err := app.usedesk.CreateTicket(newTicket)
	if err == nil {
		// Логируем успешное создание тикета
		app.Info("Создан новый тикет: ", response)

		// Сохраняем ID тикета в кэше
		err = app.ticketCache.SaveTicket(initData.User.ID, response.TicketID)
		if err != nil {
			// Логируем ошибку при сохранении тикета в кэше
			app.Error("Ошибка при сохранении тикета в кэше: ", err)
			go app.Bot.SendAllAdmins("Произошла ошибка: " + err.Error())
		}
		// Отправляем сообщение об успешном создании тикета
		go sendSuccessMessage(app, initData.User.ID, form)
	} else {
		// Логируем ошибку, если не удалось создать тикет
		app.Error("Не удалось создать запрос: ", err)
		go sendErrorMessage(app, initData.User.ID)
		go app.Bot.SendAllAdmins("Произошла ошибка: " + err.Error())
	}

	// Отправляем ответ клиенту
	app.Info("Отправка ответа клиенту: Данные успешно отправлены")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Данные успешно отправлены")
}

func sendSuccessMessage(app *WebApp, userID int64, form Form) {
	sendedText := fmt.Sprintf(`
Ваша заявка успешно отправлена.

Заявка:
	%s
	`, app.preapareMsgText(form))
	newMesg := tgbotapi.NewMessage(userID, sendedText)
	newMesg.ParseMode = "html"
	app.Bot.SendMessage(newMesg)
}

func sendErrorMessage(app *WebApp, userID int64) {
	errorText := "К сожалению, произошла ошибка при создании вашей заявки. Пожалуйста, попробуйте еще раз позже."
	newMesg := tgbotapi.NewMessage(userID, errorText)
	app.Bot.SendMessage(newMesg)
}
