package log

// CombinedLogger логгер для логирования и в консоль, и в файл
type CombinedLogger struct {
	fileLogger    *FileLogger
	consoleLogger *ConsoleLogger
}

func NewCombinedLogger(path string) (*CombinedLogger, error) {
	fileLogger, err := NewFileLogger(path)
	if err != nil {
		return nil, err
	}
	consoleLogger := NewConsoleLogger()
	return &CombinedLogger{
		fileLogger:    fileLogger,
		consoleLogger: consoleLogger,
	}, nil
}

func (c *CombinedLogger) Warn(args ...interface{}) {
	c.fileLogger.Warn(args...)
	c.consoleLogger.Warn(args...)
}

func (c *CombinedLogger) Error(args ...interface{}) {
	c.fileLogger.Error(args...)
	c.consoleLogger.Error(args...)
}

func (c *CombinedLogger) Info(args ...interface{}) {
	c.fileLogger.Info(args...)
	c.consoleLogger.Info(args...)
}

func (c *CombinedLogger) Debug(args ...interface{}) {
	c.fileLogger.Debug(args...)
	c.consoleLogger.Debug(args...)
}

func (level LogLevel) String() string {
	return string(level)
}
