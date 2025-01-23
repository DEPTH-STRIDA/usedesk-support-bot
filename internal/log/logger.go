package log

var Log Logger

// LogLevel определяет уровень логирования
type LogLevel string

const (
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
	WARN  LogLevel = "WARN"
	DEBUG LogLevel = "DEBUG"
)

// LogEntry структура для записи логов в формате JSON
type LogEntry struct {
	Timestamp string   `json:"time"`
	Level     LogLevel `json:"level"`
	Message   string   `json:"message"`
}

// Logger интерфейс для логгеров
type Logger interface {
	Error(args ...interface{})
	Warn(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
}
