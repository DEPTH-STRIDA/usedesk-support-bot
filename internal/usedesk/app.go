package usedesk

import (
	"net/http"
	"support-bot/internal/log"
	"support-bot/internal/request"
	"time"
)

const (
	baseURL = "https://api.usedesk.ru"
)

// Client представляет клиент для работы с API Usedesk
type ClientUsedesk struct {
	APIToken   string
	HTTPClient *http.Client

	requestHandler *request.RequestHandler
}

// NewClient создает новый экземпляр клиента Usedesk API
func NewClient(apiToken string) (*ClientUsedesk, error) {
	re, err := request.NewRequestHandler(log.Log, 100)
	if err != nil {
		return nil, err
	}

	go re.ProcessRequests(1 * time.Second)

	return &ClientUsedesk{
		APIToken:       apiToken,
		HTTPClient:     &http.Client{},
		requestHandler: re,
	}, nil
}
