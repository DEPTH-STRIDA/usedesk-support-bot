package usedesk

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

// NewTicketRequest представляет структуру для создания нового запроса
type NewTicketRequest struct {
	Subject      string                 `json:"subject"`
	Message      string                 `json:"message"`
	ClientName   string                 `json:"client_name,omitempty"`
	ClientEmail  string                 `json:"client_email,omitempty"`
	ClientID     int64                  `json:"client_id,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	Status       int                    `json:"status,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
	AssigneeID   int                    `json:"assignee_id,omitempty"`
	ChannelID    int                    `json:"channel_id,omitempty"`
	From         string                 `json:"from,omitempty"`
	UserID       int64                  `json:"user_id,omitempty"`
	AdditionalID string                 `json:"additional_id,omitempty"`
}

// NewTicketResponse представляет ответ на создание нового запроса
type NewTicketResponse struct {
	Status        string `json:"status"`
	TicketID      int64  `json:"ticket_id,omitempty"`
	MessageStatus string `json:"message_status,omitempty"`
	Error         string `json:"error,omitempty"`
}

// NewCommentRequest представляет структуру для создания нового комментария
type NewCommentRequest struct {
	TicketID int64  `json:"ticket_id"`
	Message  string `json:"message"`
	UserID   int64  `json:"user_id,omitempty"`
	Type     string `json:"type,omitempty"`
	From     string `json:"from,omitempty"` // user или client
}

// NewCommentResponse представляет ответ на создание нового комментария
type NewCommentResponse struct {
	Status    string `json:"status"`
	CommentID int    `json:"comment_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CreateTicket отправляет запрос на создание нового тикета
func (c *ClientUsedesk) CreateTicket(ticket NewTicketRequest) (*NewTicketResponse, error) {
	if strings.TrimSpace(ticket.Subject) == "" || strings.TrimSpace(ticket.Message) == "" {
		return nil, fmt.Errorf("предмет и сообщение обязательные для заполнения")
	}

	endpoint := fmt.Sprintf("%s/create/ticket", baseURL)

	data := url.Values{}
	data.Set("api_token", c.APIToken)
	data.Set("subject", ticket.Subject)
	data.Set("message", ticket.Message)

	if ticket.ClientName != "" {
		data.Set("client_name", ticket.ClientName)
	}
	if ticket.ClientEmail != "" {
		data.Set("client_email", ticket.ClientEmail)
	}
	if ticket.Type != "" {
		data.Set("type", ticket.Type)
	}
	if ticket.Priority != "" {
		data.Set("priority", ticket.Priority)
	}
	if ticket.Status != 0 {
		data.Set("status", strconv.Itoa(ticket.Status))
	}
	if len(ticket.Tags) > 0 {
		tagsJSON, _ := json.Marshal(ticket.Tags)
		data.Set("tags", string(tagsJSON))
	}
	if len(ticket.CustomFields) > 0 {
		customFieldsJSON, _ := json.Marshal(ticket.CustomFields)
		data.Set("custom_fields", string(customFieldsJSON))
	}
	if ticket.AssigneeID != 0 {
		data.Set("assignee_id", strconv.Itoa(ticket.AssigneeID))
	}
	if ticket.ChannelID != 0 {
		data.Set("channel_id", strconv.Itoa(ticket.ChannelID))
	}
	if ticket.From != "" {
		data.Set("from", ticket.From)
	}
	if ticket.UserID != 0 {
		data.Set("user_id", strconv.FormatInt(ticket.UserID, 10))
	}
	if ticket.ClientID != 0 {
		data.Set("client_id", strconv.FormatInt(ticket.ClientID, 10))
	}
	if ticket.AdditionalID != "" {
		data.Set("additional_id", ticket.AdditionalID)
	}

	c.requestHandler.Info(fmt.Sprintf("Отправка запроса на создание тикета: %+v", data))

	resp, err := c.HTTPClient.PostForm(endpoint, data)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при отправке запроса: %v", err))
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при чтении ответа: %v", err))
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	c.requestHandler.Info(fmt.Sprintf("Получен ответ: %s", string(body)))

	var ticketResp NewTicketResponse
	err = json.Unmarshal(body, &ticketResp)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при разборе JSON ответа: %v", err))
		return nil, fmt.Errorf("ошибка при разборе JSON ответа: %v", err)
	}

	if ticketResp.Status != "success" {
		c.requestHandler.Error(fmt.Sprintf("Ошибка создания тикета: %s", ticketResp.Error))
		return &ticketResp, fmt.Errorf("ошибка создания тикета: %s", ticketResp.Error)
	}

	c.requestHandler.Info(fmt.Sprintf("Тикет успешно создан: %+v", ticketResp))

	return &ticketResp, nil
}

// CreateComment отправляет запрос на создание нового комментария
func (c *ClientUsedesk) CreateComment(comment NewCommentRequest) (*NewCommentResponse, error) {
	if comment.TicketID == 0 || strings.TrimSpace(comment.Message) == "" {
		return nil, fmt.Errorf("id тикета и сообщение обязательные для заполнения")
	}

	endpoint := fmt.Sprintf("%s/create/comment", baseURL)

	data := url.Values{}
	data.Set("api_token", c.APIToken)
	data.Set("ticket_id", strconv.FormatInt(comment.TicketID, 10))
	data.Set("message", comment.Message)

	if comment.UserID != 0 {
		data.Set("user_id", strconv.FormatInt(comment.UserID, 10))
	}
	if comment.Type != "" {
		data.Set("type", comment.Type)
	}
	if comment.From != "" {
		data.Set("from", comment.From)
	}

	c.requestHandler.Info(fmt.Sprintf("Отправка запроса на создание комментария: %+v", data))

	resp, err := c.HTTPClient.PostForm(endpoint, data)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при отправке запроса: %v", err))
		return nil, fmt.Errorf("ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при чтении ответа: %v", err))
		return nil, fmt.Errorf("ошибка при чтении ответа: %v", err)
	}

	c.requestHandler.Info(fmt.Sprintf("Получен ответ: %s", string(body)))

	var commentResp NewCommentResponse
	err = json.Unmarshal(body, &commentResp)
	if err != nil {
		c.requestHandler.Error(fmt.Sprintf("Ошибка при разборе JSON ответа: %v", err))
		return nil, fmt.Errorf("ошибка при разборе JSON ответа: %v", err)
	}

	if commentResp.Status != "success" {
		c.requestHandler.Error(fmt.Sprintf("Ошибка создания комментария: %s", commentResp.Error))
		return &commentResp, fmt.Errorf("ошибка создания комментария: %s", commentResp.Error)
	}

	c.requestHandler.Info(fmt.Sprintf("Комментарий успешно создан: %+v", commentResp))

	return &commentResp, nil
}
