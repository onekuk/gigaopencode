package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type Logger struct {
	level   LogLevel
	logger  *log.Logger
	logFile *os.File
}

func NewLogger(level LogLevel) *Logger {
	// Создаем директорию для логов если её нет
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
	}

	// Создаем имя файла с текущей датой и временем
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFileName := filepath.Join(logDir, fmt.Sprintf("server_%s.log", timestamp))

	// Открываем файл для логирования
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v, logging to stdout only", err)
		logFile = nil
	}

	// Настраиваем вывод в stdout и файл одновременно
	var output io.Writer
	if logFile != nil {
		output = io.MultiWriter(os.Stdout, logFile)
		log.Printf("Logging to file: %s", logFileName)
	} else {
		output = os.Stdout
	}

	return &Logger{
		level:   level,
		logger:  log.New(output, "", 0),
		logFile: logFile,
	}
}

func (l *Logger) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	var prefix string
	switch level {
	case LogLevelDebug:
		prefix = "DEBUG"
	case LogLevelInfo:
		prefix = "INFO "
	case LogLevelWarn:
		prefix = "WARN "
	case LogLevelError:
		prefix = "ERROR"
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s %s", prefix, timestamp, message)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LogLevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LogLevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LogLevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LogLevelError, format, args...)
}

func GetLogLevel() LogLevel {
	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		return LogLevelDebug
	case "INFO":
		return LogLevelInfo
	case "WARN":
		return LogLevelWarn
	case "ERROR":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}
