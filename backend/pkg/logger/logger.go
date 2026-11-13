// ===== FILE: backend/pkg/logger/logger.go =====
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	logger      *log.Logger
	currentLevel = INFO
)

func init() {
	logger = log.New(os.Stdout, "", 0)
}

func SetLevel(level Level) {
	currentLevel = level
}

func format(level string, message string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(message, args...)
	return fmt.Sprintf("[%s] %s | %s", timestamp, level, msg)
}

func Debug(message string, args ...interface{}) {
	if currentLevel <= DEBUG {
		logger.Println(format("DEBUG", message, args...))
	}
}

func Info(message string, args ...interface{}) {
	if currentLevel <= INFO {
		logger.Println(format("INFO", message, args...))
	}
}

func Warn(message string, args ...interface{}) {
	if currentLevel <= WARN {
		logger.Println(format("WARN", message, args...))
	}
}

func Error(message string, args ...interface{}) {
	if currentLevel <= ERROR {
		logger.Println(format("ERROR", message, args...))
	}
}

func Fatal(message string, args ...interface{}) {
	logger.Println(format("FATAL", message, args...))
	os.Exit(1)
}
// ===== END FILE =====
