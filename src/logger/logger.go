package logger

import (
	"fmt"
	"log"
)

// LogMessageType is the type of log message.
type LogMessageType string

// Logger is used to create log output.
type Logger struct{}

const (
	Error LogMessageType = "Error"
	Warn  LogMessageType = "Warn"
	Info  LogMessageType = "Info"
)

// LogError log the error message to terminal.
func (logger *Logger) LogError(message string) {
	logMessage(Error, message)
}

// LogWarning log the warning message to terminal.
func (logger *Logger) LogWarning(message string) {
	logMessage(Warn, message)
}

// LogInfo log the normal message to terminal.
func (logger *Logger) LogInfo(message string) {
	logMessage(Info, message)
}

func logMessage(logType LogMessageType, message string) {
	switch logType {
	case Error:
		log.Fatal(fmt.Sprintf("[Error]: %s", message))
	case Warn:
		log.Println(fmt.Sprintf("[Warn]: %s", message))
	case Info:
		log.Println(fmt.Sprintf("[Info]: %s", message))
	}
}
