package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Logger struct {
	Level  int
	writer io.Writer
	file   *os.File
}

var logger Logger

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

var levelStrings = map[int]string{
	LevelFatal: "FATAL",
	LevelError: "E",
	LevelWarn:  "W",
	LevelInfo:  "I",
	LevelDebug: "D",
}

// Configura el logger
func ConfigureLogger(filepath string, level int) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	logger = Logger{
		Level:  level,
		writer: io.MultiWriter(file, os.Stdout),
		file:   file,
	}

	return nil
}

func Fatal(format string, args ...interface{}) {
	log(logger, LevelFatal, format, args...)
	os.Exit(1)
}

func Error(format string, args ...interface{}) {
	log(logger, LevelError, format, args...)
}

func Warn(format string, args ...interface{}) {
	log(logger, LevelWarn, format, args...)
}

func Info(format string, args ...interface{}) {
	log(logger, LevelInfo, format, args...)
}

func Debug(format string, args ...interface{}) {
	log(logger, LevelDebug, format, args...)
}

// Función privada, no se usa
func log(logger Logger, level int, format string, args ...interface{}) {
	if logger.Level < level {
		return
	}

	formattedTime := time.Now().Format("02/01/2006 15:04:05")
	levelString := levelStrings[level]
	formattedMessage := fmt.Sprintf(format, args...)

	stringToPrint := fmt.Sprintf("%s [ %s ] %s\n", formattedTime, levelString, formattedMessage)

	_, err := logger.writer.Write([]byte(stringToPrint))
	if err != nil {
		fmt.Printf("Could not write to logger, this should not be happening!: %v", err)
		os.Exit(1)
	}
}
