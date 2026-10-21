package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	file *os.File
}

type LogLevel int

const (
	LOG_INFO LogLevel = iota
	LOG_WARNING
	LOG_ERROR
)

func NewLogger() (*Logger, error) {
	file, err := FileLog("logs/log.txt")
	if err != nil {
		return nil, err
	}

	return &Logger{
		file: file,
	}, nil
}

func (l Logger) log(level LogLevel, value string) {
	prefix := ""
	switch level {
	case LOG_INFO:
		prefix = "[INFO] "
	case LOG_WARNING:
		prefix = "[WARNING] "
	case LOG_ERROR:
		prefix = "[ERROR] "
	}

	timestamp := time.Now().Format(time.RFC3339)
	logString := fmt.Sprintf("%s %s %s\n", timestamp, prefix, value)
	l.file.WriteString(logString)
}

func (l Logger) Write(value string) {
	l.Info(value)
}

func (l Logger) Warn(value string) {
	l.log(LOG_WARNING, value)
}

func (l Logger) Info(value string) {
	l.log(LOG_INFO, value)
}

func (l Logger) Err(value string) {
	l.log(LOG_ERROR, value)
}

func FileLog(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600) //nolint:mnd
	if err != nil {
		return nil, fmt.Errorf("error opening file for logging: %w", err)
	}

	return f, nil
}
