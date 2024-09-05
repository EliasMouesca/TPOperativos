package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	LogLevel  int
	LogWriter io.Writer
)

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

var levelStrings = map[string]int{
	"FATAL": LevelFatal,
	"ERROR": LevelError,
	"WARN":  LevelWarn,
	"INFO":  LevelInfo,
	"DEBUG": LevelDebug,
	"TRACE": LevelTrace,
}

var levelTags = map[int]string{
	LevelFatal: "FATAL",
	LevelError: "E",
	LevelWarn:  "!",
	LevelInfo:  "i",
	LevelDebug: "-",
	LevelTrace: ".",
}

// ConfigureLogger configura el logger, cuidado porque esto leakea 1 file handle...
func ConfigureLogger(filepath string, level string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	err = SetLevel(level)
	if err != nil {
		return err
	}

	LogWriter = io.MultiWriter(file, os.Stdout)

	return nil
}

func SetLevel(level string) error {
	oldLevel := LogLevel
	var exists bool

	LogLevel, exists = levelStrings[level]
	if !exists {
		LogLevel = oldLevel
		return errors.New("'" + level + "' no es un nivel válido de loggeo")
	}

	return nil
}

func Fatal(format string, args ...interface{}) {
	log(LevelFatal, format, args...)
	os.Exit(1)
}

func Error(format string, args ...interface{}) {
	log(LevelError, format, args...)
}

func Warn(format string, args ...interface{}) {
	log(LevelWarn, format, args...)
}

func Info(format string, args ...interface{}) {
	log(LevelInfo, format, args...)
}

func Debug(format string, args ...interface{}) {
	log(LevelDebug, format, args...)
}

func Trace(format string, args ...interface{}) {
	log(LevelTrace, format, args...)
}

// Función privada, no se usa
func log(level int, format string, args ...interface{}) {
	if LogLevel < level {
		return
	}

	formattedTime := time.Now().Format("02/01/2006 15:04:05")
	levelString := levelTags[level]
	formattedMessage := fmt.Sprintf(format, args...)

	stringToPrint := fmt.Sprintf("%s [ %s ] %s\n", formattedTime, levelString, formattedMessage)

	_, err := LogWriter.Write([]byte(stringToPrint))
	if err != nil {
		fmt.Printf("Could not write to logger, this should not be happening!: %v", err)
		os.Exit(1)
	}
}
