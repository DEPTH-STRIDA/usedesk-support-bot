package web

import (
	"fmt"
	"strconv"
	"time"
)

// tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type GoogleSheetData struct {
	Tegs     []string   // Набор тегов
	Problems [][]string // Набор проблем под каждый тег
}

func (app *WebApp) UpdateCache() error {
	// Получение данных
	sheetConfig := app.Sheet.GetSheetConfig()
	// Получение количества тегов
	metaData, err := app.Sheet.ReadColumnValues(sheetConfig.SheetID, sheetConfig.DataListName, "A1:A4")
	if err != nil {
		return err
	}
	if len(metaData) < 4 || metaData == nil {
		return fmt.Errorf("получение тега не удалось. Из таблицы пришел нулевой массив")
	}
	// Кол-во тегов
	numberTegs, err := strconv.ParseInt(metaData[1], 10, 64)
	if err != nil {
		return err
	}
	// Максимальное кол-во проблем по 1 тегу
	maxProblems, err := strconv.ParseInt(metaData[3], 10, 64)
	if err != nil {
		return err
	}
	values, err := app.Sheet.ReadRangeByNumbers(sheetConfig.SheetID, sheetConfig.DataListName, 1, 2, 2+int(maxProblems), 1+int(numberTegs))
	if err != nil {
		return err
	}
	// Массив имен тегов из первой строки
	tegsNames := values[0]
	Problems := [][]string{}

	// for i := range values {
	// 	fmt.Println(i, " ", values[i])
	// }

	// Обход строки с размерами
	for i, v := range values[1] {
		// Получение количества проблем у данного тега
		currentSize, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		// Получение всех проблем по данному тегу
		problems := []string{}
		for j := 2; j < 2+int(currentSize); j++ {
			problems = append(problems, values[j][i])
		}
		Problems = append(Problems, problems)
	}
	data := GoogleSheetData{
		tegsNames,
		Problems,
	}
	app.Cache.SetNewData(data)
	return nil
}

func (app *WebApp) StartPeriodUpdateCache(pause time.Duration) {
	ticker := time.NewTicker(pause)

	// Используем канал для сигнала о завершении
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				app.Info("Обновление кеша раз в ", t, " началось")
				app.UpdateCache()
			}
		}
	}()
}
