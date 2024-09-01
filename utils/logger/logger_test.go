package logger

import (
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	ConfigureLogger("test.log", "TRACE")

	Trace("This is trace level")
	Debug("This is debug, myVar = %v", 23)
	Info("This is some useful info :p")

	Warn("This is a warning! File '%v' could not be found.", "falopa.file")

	_, err := os.Open("sas")
	Error("This is an error with file '%s' - %v", "sas", err)

	Fatal("This is a fatal error, execution will continue no longer !")

}
