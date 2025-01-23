package googlesheet

import (
	"context"
	"fmt"
	"support-bot/internal/log"
	"support-bot/internal/request"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetManager interface {
	ReadColumnValues(sheetID, listName, column string) ([]string, error)
	ReadRowValues(sheetID, listName, row string) ([]string, error)
	ReadRangeValues(sheetID, listName, rangeA1 string) ([][]string, error)
	// ReadRangeByNumbers синхронно считывает данные из заданного диапазона, используя числовые координаты
	// SheetID - id google таблицы, sheetName - имя листа
	// startRow, startCol - начальная строка и столбец (нумерация с 1)
	// endRow, endCol - конечная строка и столбец (нумерация с 1)
	ReadRangeByNumbers(sheetID, listName string, startRow, startCol, endRow, endCol int) ([][]string, error)
	GetSheetConfig() Config
}

// GoogleSheets - структуру для работы с Google таблицами
type GoogleSheets struct {
	log.Logger                              // Логгер
	*sheets.Service                         // Сервис Google Sheets API
	config          Config                  // Структура с параметрами, необходимыми для работы с таблицами
	Request         *request.RequestHandler // Структура для откладывания выполнений функций
}

// NewGoogleSheets создает новый экземпляр GoogleSheets
func NewGoogleSheets(config Config) (*GoogleSheets, error) {
	Request, err := request.NewRequestHandler(log.Log, int64(config.BufferSize))
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	service, err := sheets.NewService(ctx, option.WithCredentialsFile("config/credentials.json"), option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, fmt.Errorf("не удается инициализировать сервис Google Sheets: %v", err)
	}

	app := &GoogleSheets{
		Logger:  log.Log,
		config:  config,
		Request: Request,
		Service: service,
	}

	pause := time.Duration(app.config.RequestUpdatePause) * time.Second
	go app.Request.ProcessRequestsWithDynamicPause(pause, request.IncrementPause(1.5, 30*time.Second))

	return app, nil
}

// ReadColumnValues синхроно считывает данные из столбца
// SheetID - id google таблицы, sheetName - имя листа, column - колона (следует указывать A:A или A1:A, A2:A, если надо пропусти первый)
func (app *GoogleSheets) ReadColumnValues(sheetID, listName, column string) ([]string, error) {
	var values []string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	app.Request.HandleRequest(func() error {
		defer wg.Done()
		readRange := fmt.Sprintf("%s!%s", listName, column)
		var resp *sheets.ValueRange
		resp, err = app.Service.Spreadsheets.Values.Get(sheetID, readRange).Do()
		if err != nil {
			return fmt.Errorf("не удалось извлечь данные из таблицы: %v", err)
		}
		for _, row := range resp.Values {
			var rowData []string
			for _, cell := range row {
				cellValue, ok := cell.(string)
				if !ok {
					return fmt.Errorf("неожиданный тип значения в таблице")
				}
				rowData = append(rowData, cellValue)
			}
			if len(rowData) > 0 {
				values = append(values, rowData[0])
			}
		}
		return nil
	})

	wg.Wait()
	return values, err
}

// ReadRowValues синхронно считывает данные из строки
// SheetID - id google таблицы, sheetName - имя листа, row - строка (следует указывать 1:1 или A1:Z1, B1:Z1, если надо пропустить первый столбец)
func (app *GoogleSheets) ReadRowValues(sheetID, listName, row string) ([]string, error) {
	var values []string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	app.Request.HandleRequest(func() error {
		defer wg.Done()
		readRange := fmt.Sprintf("%s!%s", listName, row)
		var resp *sheets.ValueRange
		resp, err = app.Service.Spreadsheets.Values.Get(sheetID, readRange).Do()
		if err != nil {
			return fmt.Errorf("не удалось извлечь данные из таблицы: %v", err)
		}
		if len(resp.Values) > 0 {
			for _, cell := range resp.Values[0] {
				cellValue, ok := cell.(string)
				if !ok {
					return fmt.Errorf("неожиданный тип значения в таблице")
				}
				values = append(values, cellValue)
			}
		}
		return nil
	})

	wg.Wait()
	return values, err
}

// ReadRangeValues синхронно считывает данные из заданного диапазона
// SheetID - id google таблицы, sheetName - имя листа, rangeA1 - диапазон в нотации A1 (например, "A1:D12")
func (app *GoogleSheets) ReadRangeValues(sheetID, listName, rangeA1 string) ([][]string, error) {
	var values [][]string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	app.Request.HandleRequest(func() error {
		defer wg.Done()
		readRange := fmt.Sprintf("%s!%s", listName, rangeA1)
		var resp *sheets.ValueRange
		resp, err = app.Service.Spreadsheets.Values.Get(sheetID, readRange).Do()
		if err != nil {
			return fmt.Errorf("не удалось извлечь данные из таблицы: %v", err)
		}

		values = make([][]string, len(resp.Values))
		for i, row := range resp.Values {
			values[i] = make([]string, len(row))
			for j, cell := range row {
				cellValue, ok := cell.(string)
				if !ok {
					// Если значение не является строкой, преобразуем его в строку
					values[i][j] = fmt.Sprintf("%v", cell)
				} else {
					values[i][j] = cellValue
				}
			}
		}
		return nil
	})

	wg.Wait()
	return values, err
}

// columnToLetter преобразует числовой индекс столбца в буквенное обозначение
func columnToLetter(column int) string {
	var result string
	for column > 0 {
		column--
		result = string(rune('A'+column%26)) + result
		column /= 26
	}
	return result
}

// ReadRangeByNumbers синхронно считывает данные из заданного диапазона, используя числовые координаты
// SheetID - id google таблицы, sheetName - имя листа
// startRow, startCol - начальная строка и столбец (нумерация с 1)
// endRow, endCol - конечная строка и столбец (нумерация с 1)
func (app *GoogleSheets) ReadRangeByNumbers(sheetID, listName string, startRow, startCol, endRow, endCol int) ([][]string, error) {
	var values [][]string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)

	app.Request.HandleRequest(func() error {
		defer wg.Done()

		// Преобразуем числовые координаты в нотацию A1
		startColumn := columnToLetter(startCol)
		endColumn := columnToLetter(endCol)
		rangeA1 := fmt.Sprintf("%s%d:%s%d", startColumn, startRow, endColumn, endRow)

		readRange := fmt.Sprintf("%s!%s", listName, rangeA1)
		var resp *sheets.ValueRange
		resp, err = app.Service.Spreadsheets.Values.Get(sheetID, readRange).Do()
		if err != nil {
			return fmt.Errorf("не удалось извлечь данные из таблицы: %v", err)
		}

		values = make([][]string, len(resp.Values))
		for i, row := range resp.Values {
			values[i] = make([]string, len(row))
			for j, cell := range row {
				cellValue, ok := cell.(string)
				if !ok {
					// Если значение не является строкой, преобразуем его в строку
					values[i][j] = fmt.Sprintf("%v", cell)
				} else {
					values[i][j] = cellValue
				}
			}
		}
		return nil
	})

	wg.Wait()
	return values, err
}

func (app *GoogleSheets) GetSheetConfig() Config {
	return app.config
}
