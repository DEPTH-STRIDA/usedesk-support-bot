package tg

import (
	"fmt"
	"support-bot/internal/log"
	"support-bot/internal/usedesk"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HandlerFunc func(app *Bot, update tgbotapi.Update) error

const (
	startMsg0 = `
Приветствую!
В этого бота через форму отправляются запросы, касающиеся учеников и групп, связанные с технической частью, расписанием и сопровождением
Нажмите слева на кнопку «Запрос» чтобы открылась форма, которую вам необходимо будет заполнить! Отправленный запрос попадает в Юздеск. Ответ на ваш запрос появится в этом чате.
 Укажите максимально подробно вашу проблему, чтобы сервис смог ее решить. Пишите вежливо и с уважением к другим сотрудникам 😌
Режим работы отделов: с 9:00 до 21:00 мск
Если вдруг форма не работает, или у вас есть вопросы/предложение/баги/проблемы - пишите @slice13

`
	startMsg1 = `
Правила работы с ботом:
 💫 запросы должны быть строго по формату
 💫 1 клиент - 1 запрос
Бот отправляет запросы в Юздеск.

`
	startMsg2Pined = `
Подробное описание к какому отделу обращаться, в случае проблемы:
#техи
1. Отметка явки учеников в crm
2. Поиск учеников в начале урока (если в вашем чате не отвечают более 10 минут)
3. Помощь в подключении к уроку, вылеты с уроков
4. Отправка записей, дз, материалов по запросу
5. Помощь в тех проблемах с микро, звуком, демонстрацией
6. Помощь в переустановке программ (не разбираемся с кодом и настройками внутри программ)
7. Перенос курсов и учеников на платформу EasyCode
8. Смена препода на платформах EasyCode, EasyEng
9. Проблемы с life digital

#расписание
1. Дубли в расписании
2. Наложение уроков
3. Исправление дат в курсе
4. Продление курсов
5. Консолидация/расформирование группы
6. Неправильно определен уровень ученика

 #забота
1. Дисциплина ребенка
2. Назначить отработку
3. Отказывались на уроке
4. Проблемы с посещаемостью ребенка
5. Перезапись на другое направление/группу /формат
6. Связь с родителями
7. Перенос ПОДАРОЧНОГО индивидуального обучения
8. ОС по урокам
`
)

func handleError(err error) error {
	log.Log.Error(err.Error())
	return err
}

func (bot *Bot) handleErrorWithAdmins(err error) error {
	go bot.SendAllAdmins(err.Error())
	return err
}

// handleStartMessage пересылает пользователю стикер в ответ на начало работы
func HandleStartMessage() HandlerFunc {
	return func(app *Bot, update tgbotapi.Update) error {

		newMsg := tgbotapi.NewMessage(update.Message.From.ID, startMsg0)
		newMsg.ParseMode = "html"
		_, err := app.SendMessage(newMsg)
		if err != nil {
			handleError(err)
		}

		newMsg = tgbotapi.NewMessage(update.Message.From.ID, startMsg1)
		newMsg.ParseMode = "html"
		_, err = app.SendMessage(newMsg)
		if err != nil {
			handleError(err)
		}

		// Отправка соощения, которое будет прикреплено
		msgText := tgbotapi.NewMessage(update.Message.From.ID, startMsg2Pined)
		msgText.ParseMode = "html"
		pinupMsg, err := app.SendMessage(msgText)
		if err != nil {
			return handleError(err)
		}

		// Открепление всех закрепов
		unpinConfig := tgbotapi.UnpinAllChatMessagesConfig{
			ChatID:          update.Message.Chat.ID,
			ChannelUsername: update.Message.From.UserName,
		}
		_, err = app.BotAPI.Request(unpinConfig)
		if err != nil {
			return handleError(err)
		}

		// Закрепление отправленного сообщения
		pinConfig := tgbotapi.PinChatMessageConfig{
			ChatID:              update.Message.Chat.ID,
			MessageID:           pinupMsg.MessageID,
			DisableNotification: false, // Если true, уведомление о закреплении не будет отправлено
		}
		_, err = app.BotAPI.Request(pinConfig)
		if err != nil {
			return handleError(err)
		}
		return nil
	}
}

// handleStartMessage отправляет пользователю сообщение с кнопкой под ним. Кнопка содержит ссылку.
func HandleSendFormMessage(wepAppUrl, msgText, buttonMsg string) HandlerFunc {
	return func(app *Bot, update tgbotapi.Update) error {
		// Создание сообщения с кнопкой
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		// Создание кнопки
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonURL(buttonMsg, wepAppUrl),
		}
		// Установка кнопки в сообщение
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(row)

		// Отправка сообщения
		_, err := app.sendMessage(msg)
		if err != nil {
			handleError(err)
		}
		return nil
	}
}

func (app *Bot) HandleUserMessage(update tgbotapi.Update) error {
	if update.Message == nil {
		return handleError(fmt.Errorf("message is nil in HandleUserMessage"))
	}

	if update.Message.From == nil {
		return handleError(fmt.Errorf("message.From is nil in HandleUserMessage"))
	}

	// Проверка зависимостей
	if app.ticketCache == nil {
		return handleError(fmt.Errorf("ticketCache is nil in HandleUserMessage"))
	}
	if app.usedesk == nil {
		return handleError(fmt.Errorf("usedesk is nil in HandleUserMessage"))
	}

	// Логирование получения текущего тикета ID
	log.Log.Info("Получение текущего тикета ID для пользователя с Telegram ID: ", update.Message.From.ID)

	// Получение текущего тикета ID
	telegramID := update.Message.From.ID
	ticketID, ok := app.ticketCache.GetCurrentTicketIDByTgId(telegramID)
	// Если тикет получен, то отправляем запрос с комментарием
	if ok {
		return app.createUsedeskCommentary(update, ticketID)
		// Если не удалось получить тикет, значит надо создать новый тикет с чатом
	} else {
		return app.createUsedeskTickerChat(update)
	}
}

func (app *Bot) createUsedeskCommentary(update tgbotapi.Update, ticketID int64) error {
	// Логирование создания комментария
	log.Log.Info("Создание комментария для тикета ID: ", ticketID)

	if ticketID == 0 {
		log.Log.Warn("Попытка создать комментарий для тикета с ID 0")
		return app.handleErrorWithAdmins(fmt.Errorf("ошибка при создании комментария в usedesk: ticketID=0"))
	}

	comm := usedesk.NewCommentRequest{
		Message:  update.Message.Text,
		TicketID: ticketID,
		Type:     "public",
		From:     "client",
	}
	_, err := app.usedesk.CreateComment(comm)

	if err != nil {
		log.Log.Info("Ошибка при создании комментария в usedesk: ", err)
		return app.handleErrorWithAdmins(fmt.Errorf("ошибка при создании комментария в usedesk: %v", err))
	}

	return nil
}

func (app *Bot) createUsedeskTickerChat(update tgbotapi.Update) error {
	// Логирование создания нового тикета
	log.Log.Info("Создание нового тикета для пользователя: ", update.Message.From.FirstName, " ", update.Message.From.LastName)

	tick := usedesk.NewTicketRequest{
		Message:    update.Message.Text,
		Subject:    "Telegram chat",
		ClientName: update.Message.From.FirstName + " " + update.Message.From.LastName + "(" + update.Message.From.UserName + ")",
	}

	resp, err := app.usedesk.CreateTicket(tick)
	if err != nil {
		log.Log.Info("Ошибка при создании тикета: ", err)
		return app.handleErrorWithAdmins(fmt.Errorf("ошибка при создании тикета: %v", err))
	}

	log.Log.Info("Тикет успешно создан с ID: ", resp.TicketID)

	err = app.ticketCache.SaveTicket(update.Message.From.ID, resp.TicketID)
	if err != nil {
		log.Log.Info("Ошибка при сохранении тикета: ", err)
		return app.handleErrorWithAdmins(fmt.Errorf("ошибка при сохранении тикета: %v", err))
	}
	return nil
}
