package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileLogger логгер для логирования в один файл
type FileLogger struct {
	mu        sync.Mutex
	file      *os.File
	path      string
	startTime time.Time
}

// NewFileLogger конструктор файлового логгера
func NewFileLogger(basePath string) (*FileLogger, error) {
	logger := &FileLogger{
		path:      basePath,
		startTime: time.Now(),
	}

	err := logger.initialize()
	if err != nil {
		return nil, err
	}

	go logger.rotateLogs()

	return logger, nil
}

// initialize создает файл для логирования
func (f *FileLogger) initialize() error {
	dir := filepath.Dir(f.path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	filename := filepath.Join(dir, "log_"+time.Now().Format("02-01-2006_15-04-05")+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

// rotateLogs отвечает за периодическое обновление лог-файла
func (f *FileLogger) rotateLogs() {
	for {
		time.Sleep(time.Until(f.startTime.AddDate(0, 1, 0)))
		f.mu.Lock()
		f.file.Close()
		f.file = nil
		f.startTime = time.Now()
		err := f.initialize()
		if err != nil {
			fmt.Printf("Failed to rotate logs: %v\n", err)
		}
		f.mu.Unlock()
	}
}

func (f *FileLogger) log(level LogLevel, args ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format("02-01-2006 15:04:05"),
		Level:     level,
		Message:   fmt.Sprint(args...),
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintln(f.file, string(data))
}

func (f *FileLogger) Warn(args ...interface{}) {
	f.log(WARN, args...)
}

func (f *FileLogger) Error(args ...interface{}) {
	f.log(ERROR, args...)
}

func (f *FileLogger) Info(args ...interface{}) {
	f.log(INFO, args...)
}

func (f *FileLogger) Debug(args ...interface{}) {
	f.log(DEBUG, args...)
}
