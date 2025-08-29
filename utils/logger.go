package utils

import (
	"discord-bot/constants"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var globalLogger *Logger

func init() {
	globalLogger = NewLogger()
}

func NewLogger() *Logger {
	level := getLogLevelFromEnv()
	logger := log.New(os.Stdout, "", 0)

	return &Logger{
		level:  level,
		logger: logger,
	}
}

func getLogLevelFromEnv() LogLevel {
	levelStr := strings.ToUpper(os.Getenv(constants.EnvLogLevel))
	switch levelStr {
	case constants.LogLevelDebug:
		return DEBUG
	case constants.LogLevelInfo:
		return INFO
	case constants.LogLevelWarn:
		return WARN
	case constants.LogLevelError:
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	levelStr := l.getLevelString(level)
	timestamp := time.Now().Format(constants.DateTimeFormat)
	message := fmt.Sprintf(format, args...)

	l.logger.Printf("[%s] %s %s", timestamp, levelStr, message)
}

func (l *Logger) getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return constants.LogLevelDebug
	case INFO:
		return constants.LogLevelInfo
	case WARN:
		return constants.LogLevelWarn
	case ERROR:
		return constants.LogLevelError
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// 글로벌 로거 함수들
func Debug(format string, args ...interface{}) {
	globalLogger.Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	globalLogger.Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Error(format, args...)
}
