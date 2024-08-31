package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	log, _ := CreateLogger("test", LevelDebug)

	log.Debug("This is debug, myVar = %v", 23)
	log.Info("This is some useful info :p")

	log.Warn("This is a warning! File '%v' could not be found.", "falopa.file")

	_, err := os.Open("sas")
	log.Error("This is an error with file '%s' - %v", "sas", err)

	log.Fatal("This is a fatal error, execution will continue no longer !")

}
