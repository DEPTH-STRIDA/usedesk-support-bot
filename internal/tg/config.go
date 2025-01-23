package tg

type Config struct {
	Token string `json:"token"`

	SupportChatID   int64  `json:"support-chat-ID"`
	SupportChatName string `json:"support-chat-name"`
	StartStickerID  string `json:"start-sticker-ID"` // Стикер, который будет отправляться в ответ на /start

	Admins []int64 `json:"admins"` // Список телеграм ID's админов телеграмм

	RequestUpdatePause         int `json:"request-update-pause"`          // Время между выполнениями запросов
	RequestCallBackUpdatePause int `json:"request-callback-update-pause"` // Пауза между обработками callback
	MsgBufferSize              int `json:"message-buffer-size"`           // Максимальное возможно количество отложенных запросов для канала сообщений
	CallBackBufferSize         int `json:"callback-buffer-size"`          // Максимальное возможно количество отложенных запросов для канала callback событий

	BotTgChat    int64 `json:"bot-tg-chat"`
	ErrorTopicID int   `json:"error-topic-id"` // ID сообщенияв беседе бот тг. Нужно для пересылки в тему.
}
