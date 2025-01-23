package tg

import (
	"errors"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обратный отсчет перед отправкой медиагруппы. Если в течении этого времени, приходит еще член медиагруппы, то отчет начинается заного
var pauseSecond time.Duration = 5

// CountdownTimer структура для управления обратным отсчетом.
type CountdownTimer struct {
	Duration  time.Duration
	IsRunning bool
	timer     *time.Timer
	onFinish  func()
}

// NewCountdownTimer создает новый обратный отсчет с указанной продолжительностью.
func NewCountdownTimer(duration time.Duration) *CountdownTimer {
	return &CountdownTimer{
		Duration:  duration,
		IsRunning: false,
	}
}

// Start начинает обратный отсчет и запускает указанную функцию по истечении времени.
func (ct *CountdownTimer) Start(onFinish func()) {
	ct.IsRunning = true
	ct.onFinish = onFinish
	ct.timer = time.AfterFunc(ct.Duration, ct.onFinish)
}

// Reset сбрасывает обратный отсчет.
func (ct *CountdownTimer) Reset() {
	if ct.timer != nil {
		ct.timer.Stop()
		ct.IsRunning = false
	}
	ct.Start(ct.onFinish)
}

// Структура для хранения медиагруппы
type MediaGroup struct {
	ChatID         int64             // ID отправителя
	Name           string            // Имя отправителя
	MediaGroupID   string            // ID медиагруппы
	Media          []tgbotapi.FileID // Список членов медиагруппы
	Caption        []string          // Подпись к медиагруппе
	Type           []string          // Тип медиагруппы (фото, видео и т.д.)
	CountdownTimer                   //Обратный отсчет, запускающий функцию
}

// Мьютекс для контроля доуступа к медиагруппам
var Mutex sync.Mutex

// Карта для хранения медиагрупп по их ID
var mediaGroups map[string]*MediaGroup = make(map[string]*MediaGroup)

// Функция для обработки сообщений входящих в медиагруппу
func (app *Bot) handleMediaGroup(update tgbotapi.Update) error {
	app.Info("Обработка медиагруппы от: ", update.Message.Chat.ID)
	//id медиагруппы
	mediaGroupID := update.Message.MediaGroupID

	//Проверка
	if mediaGroupID == "" {
		return errors.New("полученное сообщение не является членом медиагруппы")
	}
	if update.Message.Chat.ID == app.config.BotTgChat || update.Message.Chat.ID == app.config.SupportChatID {
		app.Error("Игнорирование обновления из главных чатов: ", update.Message.Chat.ID)
		return nil
	}

	//Определяем тип файла и его id
	var typeMess string
	var fileId tgbotapi.FileID
	switch {
	case len(update.Message.Photo) > 0:
		typeMess = "photo"
		fileId = tgbotapi.FileID(update.Message.Photo[len(update.Message.Photo)-1].FileID)
	case update.Message.Audio != nil:
		typeMess = "audio"
		fileId = tgbotapi.FileID(update.Message.Audio.FileID)

	case update.Message.Document != nil:
		typeMess = "document"
		fileId = tgbotapi.FileID(update.Message.Document.FileID)

	case update.Message.Video != nil:
		typeMess = "video"
		fileId = tgbotapi.FileID(update.Message.Video.FileID)
	}

	//Перед проверкой медиагруппы, включаем мьютекс
	Mutex.Lock()
	app.Info("Обработка члена медиагруппы " + typeMess + " (" + mediaGroupID + ")")
	// Проверяем наличие медиагруппы с указанным ID
	// Если группы не было, то запускаем таймер по истечении, которого отправится медиагруппа
	group, exists := mediaGroups[mediaGroupID]
	if !exists {
		// Если медиагруппы с таким ID нет, создаем новую
		group = &MediaGroup{
			ChatID:       update.Message.Chat.ID,
			Name:         update.Message.Chat.UserName,
			MediaGroupID: mediaGroupID,
			Media:        []tgbotapi.FileID{fileId},
			Caption:      []string{update.Message.Caption},
			Type:         []string{typeMess},
		}
		mediaGroups[mediaGroupID] = group
		mediaGroups[mediaGroupID].CountdownTimer = *NewCountdownTimer(pauseSecond * time.Second)
		go mediaGroups[mediaGroupID].Start(func() {
			app.SendMediaGroup(mediaGroups[mediaGroupID])
		})
		Mutex.Unlock()
		return nil
	}

	// Добавляем новый член медиагруппы
	group.Media = append(group.Media, fileId)
	group.Caption = append(group.Caption, update.Message.Caption)
	group.Type = append(group.Type, typeMess)
	group.Reset()
	Mutex.Unlock()
	return nil
}

// Функция для сборки медиагруппы и отправки ее в чат при необходимости
func (app *Bot) SendMediaGroup(group *MediaGroup) {
	defer Mutex.Unlock()
	// Проверяем, прошло ли более 5 секунд с последнего обновления медиагруппы
	Mutex.Lock()

	// Формируем сообщение с медиагруппой
	mediaGroup := []interface{}{}
	for i, v := range group.Media {
		switch group.Type[i] {
		case "photo":
			file := tgbotapi.NewInputMediaPhoto(v)
			file.Caption = group.Caption[i]
			mediaGroup = append(mediaGroup, file)
		case "audio":
			file := tgbotapi.NewInputMediaAudio(v)
			file.Caption = group.Caption[i]
			mediaGroup = append(mediaGroup, file)

		case "document":
			file := tgbotapi.NewInputMediaDocument(v)
			file.Caption = group.Caption[i]
			mediaGroup = append(mediaGroup, file)

		case "video":
			file := tgbotapi.NewInputMediaVideo(v)
			file.Caption = group.Caption[i]
			mediaGroup = append(mediaGroup, file)
		default:
			return
		}
	}
	msg := tgbotapi.NewMediaGroup(app.GetConfig().SupportChatID, mediaGroup)

	// Отправляем медиагруппу в чат

	var sendedMsg []tgbotapi.Message
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	app.msgRequestHandler.HandleRequest(func() error {
		defer wg.Done()
		sendedMsg, err = app.BotAPI.SendMediaGroup(msg)
		return nil
	})

	wg.Wait()

	if err != nil {
		app.Error("Ошибка при отправке медиагруппы в чат: " + err.Error())
		return
	}

	if len(sendedMsg) <= 0 {
		return
	}

	url := fmt.Sprintf("https://t.me/c/%s/%d", app.GetConfig().SupportChatName, sendedMsg[len(sendedMsg)-1].MessageID)
	sendedText := fmt.Sprintf(`
Ваша заявка успешно отправлена.
Заявка в чате поддержки: %s
	`, url)

	newMesg := tgbotapi.NewMessage(group.ChatID, sendedText)
	newMesg.ParseMode = "html"
	app.SendMessage(newMesg)

	app.Info("Сообщение успешно перенаправлено в чат (от: " + group.Name + "; тип: mediagroup-" + group.Type[0])
	delete(mediaGroups, group.MediaGroupID)
}
