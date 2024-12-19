package logger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	LogLevel      int
	FileWriter    io.Writer
	ConsoleWriter io.Writer
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
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	err = SetLevel(level)
	if err != nil {
		return err
	}

	FileWriter = file
	ConsoleWriter = os.Stdout

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

	//formattedTime := time.Now().Format("02/01/2006 15:04:05")
	formattedTime := time.Now().Format("15:04:05.000")
	levelString := levelTags[level]
	formattedMessage := fmt.Sprintf(format, args...)

	// Get line and file
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		fmt.Println("Unable to retrieve caller information")
		return
	}

	stringToFile := fmt.Sprintf("%s [ %s ] %s \t ->%s:%d\n",
		formattedTime, levelString, formattedMessage, filepath.Base(file), line)

	stringColored := fmt.Sprintf("%s%v%s [ %s ] %v%s -> %s:%d%s\n",
		escapeSequences["grey"], formattedTime,
		escapeSequences[levelColors[level]],
		levelString, formattedMessage,
		escapeSequences["grey"],
		filepath.Base(file), line,
		escapeSequences["reset"])

	_, err := FileWriter.Write([]byte(stringToFile))
	if err != nil {
		fmt.Printf("Could not write log to file, this should not be happening!: %v", err)
		os.Exit(1)
	}

	_, err = ConsoleWriter.Write([]byte(stringColored))
	if err != nil {
		fmt.Printf("Could not write log to console, this should not be happening!: %v", err)
		os.Exit(1)
	}

}

var escapeSequences = map[string]string{
	"reset":      "\033[0m",
	"bold":       "\033[1m",
	"grey":       "\033[90m",
	"red":        "\033[31m",
	"bright_red": "\033[91m",
	"green":      "\033[32m",
	"yellow":     "\033[33m",
	"blue":       "\033[34m",
	"cyan":       "\033[35m",
	"white":      "\033[0m",
}

var levelColors = map[int]string{
	LevelFatal: "bright_red",
	LevelError: "red",
	LevelWarn:  "yellow",
	LevelInfo:  "green",
	LevelDebug: "reset",
	LevelTrace: "grey",
}
