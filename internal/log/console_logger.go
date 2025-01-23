package log

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ConsoleLogger логгер для логирования в консоль
type ConsoleLogger struct {
	mu sync.Mutex
}

// NewConsoleLogger конструктор консольного логгера
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (c *ConsoleLogger) log(level LogLevel, args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format("02-01-2006 15:04:05"),
		Level:     level,
		Message:   fmt.Sprint(args...),
	}
	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

func (c *ConsoleLogger) Warn(args ...interface{}) {
	c.log(WARN, args...)
}

func (c *ConsoleLogger) Error(args ...interface{}) {
	c.log(ERROR, args...)
}

func (c *ConsoleLogger) Info(args ...interface{}) {
	c.log(INFO, args...)
}

func (c *ConsoleLogger) Debug(args ...interface{}) {
	c.log(DEBUG, args...)
}
