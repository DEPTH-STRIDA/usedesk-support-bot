package googlesheet

type Config struct {
	SheetID string `json:"sheet-id"`

	DataListName string `json:"data-list-name"`

	RequestUpdatePause  int `json:"request-update-pause"`   // Время между выполнениями запросов
	DataUpdatePauseHour int `json:"data-update-pause-hour"` // Время между обновлениями данных выпадающих списков
	BufferSize          int `json:"buffer-size"`            // Максимальное возможно количество отложенных запросов
}
