package web

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
	"time"
)

// NewTemplateCache находит файлы шаблонов и создает картку маршрутов
func NewTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	files, err := filepath.Glob(filepath.Join(dir, "*.tmpl"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		match, err := filepath.Match("*.page.tmpl", filepath.Base(file))
		if err != nil {
			return nil, err
		}
		if match {
			name := filepath.Base(file)
			ts, err := template.ParseFiles(file)
			if err != nil {
				return nil, err
			}
			cache[name] = ts
		}
	}
	return cache, nil
}

func (app *WebApp) render(w http.ResponseWriter, name string, data map[string]interface{}) error {
	ts, ok := app.TemplateCache[name]
	if !ok {
		return errors.New("не удалось использовать шаблон. Шаблон " + name + " не существует")
	}

	// Если data не предоставлена, создаем пустую map
	if data == nil {
		data = make(map[string]interface{})
	}

	// Добавляем версию, если она еще не установлена
	if _, exists := data["Version"]; !exists {
		data["Version"] = fmt.Sprintf("%d", time.Now().Unix())
	}

	err := ts.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}
