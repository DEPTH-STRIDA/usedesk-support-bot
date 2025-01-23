package usedesk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Event базовая структура события
type Event struct {
	Type    string
	Payload interface{}
}

// MessageEvent структура для сообщений от агента через форму
type MessageEvent struct {
	Message  string
	UserID   int64
	UserName string
	TicketID int64
}

// CommentEvent структура для комментариев
type CommentEvent struct {
	Comment struct {
		TicketID int64
		From     string
		Message  string
		Type     string
	}
}

// TriggerEvent структура для триггеров
type TriggerEvent struct {
	Trigger struct {
		TicketID  int64
		NewStatus string
	}
}

// Структуры для парсинга JSON
type rawCommentJSON struct {
	Comment struct {
		TicketID int64  `json:"ticket_id"`
		From     string `json:"from"`
		Message  string `json:"message"`
		Type     string `json:"type"`
	} `json:"comment"`
}

type rawTriggerEvent struct {
	Trigger struct {
		TicketID  int64       `json:"ticket_id"`
		NewStatus json.Number `json:"new_status"`
	} `json:"trigger"`
}

// DetermineEventType определяет тип события из запроса
func DetermineEventType(r *http.Request) (Event, error) {
	contentType := r.Header.Get("Content-Type")

	// Обработка JSON запросов
	if strings.Contains(contentType, "application/json") {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return Event{}, fmt.Errorf("error reading body: %v", err)
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// Пробуем распарсить как комментарий
		var rawComment rawCommentJSON
		if err := json.Unmarshal(body, &rawComment); err == nil &&
			rawComment.Comment.TicketID != 0 {

			event := CommentEvent{
				Comment: struct {
					TicketID int64
					From     string
					Message  string
					Type     string
				}{
					TicketID: rawComment.Comment.TicketID,
					From:     rawComment.Comment.From,
					Message:  rawComment.Comment.Message,
					Type:     rawComment.Comment.Type,
				},
			}
			return Event{Type: "comment", Payload: event}, nil
		}

		// Пробуем распарсить как триггер
		var rawEvent rawTriggerEvent
		if err := json.Unmarshal(body, &rawEvent); err == nil && rawEvent.Trigger.TicketID != 0 {
			event := TriggerEvent{
				Trigger: struct {
					TicketID  int64
					NewStatus string
				}{
					TicketID:  rawEvent.Trigger.TicketID,
					NewStatus: string(rawEvent.Trigger.NewStatus),
				},
			}
			return Event{Type: "trigger", Payload: event}, nil
		}

		return Event{}, fmt.Errorf("unknown JSON event type")
	}

	// Обработка multipart/form-data запросов
	if strings.Contains(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return Event{}, fmt.Errorf("error parsing multipart form: %v", err)
		}

		// Обработка сообщения от агента
		if message := r.FormValue("message"); message != "" {
			userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
			ticketID, _ := strconv.ParseInt(r.FormValue("ticket_id"), 10, 64)

			event := MessageEvent{
				Message:  message,
				UserID:   userID,
				UserName: r.FormValue("user_name"),
				TicketID: ticketID,
			}
			return Event{Type: "message", Payload: event}, nil
		}

		// Обработка комментария через форму
		if comment := r.FormValue("comment"); comment != "" {
			ticketID, _ := strconv.ParseInt(r.FormValue("ticket_id"), 10, 64)

			event := CommentEvent{
				Comment: struct {
					TicketID int64
					From     string
					Message  string
					Type     string
				}{
					TicketID: ticketID,
					From:     r.FormValue("from"),
					Message:  r.FormValue("message"),
					Type:     r.FormValue("type"),
				},
			}
			return Event{Type: "comment", Payload: event}, nil
		}
	}

	return Event{}, fmt.Errorf("unsupported Content-Type: %s", contentType)
}
