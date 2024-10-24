package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var config fsConfig

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("filesystem.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}

	data, err := os.ReadFile("config.json")
	if err != nil {
		logger.Error("No se pudo leer la config - ", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		logger.Error("Error parseando la config - ", err)
	}

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Error("Error seteando el log level - ", err)
	}

}

func main() {
	dino.Pterodactyl()
	logger.Info("--- Comienzo ejecución del filesystem ---")

	var err error

	http.HandleFunc("/", notFound)
	http.HandleFunc("POST /memoryDump", persistMemoryDump)

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Debug("Corriendo filesystem en %v", self)
	err = http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("ListenAndServe terminó con un error - %v", err)
	}
}

func assertBitmapExists() error {
	filename := "bitmap.dat"
	_, err := os.Stat(filename)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

	}

	return nil
}

func assertBloquesExists() error {

	return nil
}

func notFound(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request inválida %v, desde %v", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request inválida"))
	if err != nil {
		logger.Error("No se pudo escribir la respuesta - %v", err.Error())
	}
}
