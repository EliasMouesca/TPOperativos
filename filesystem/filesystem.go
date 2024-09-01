package main

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"log"
	"net/http"
)

func init() {
	loggerLevel := logger.LevelInfo
	err := logger.ConfigureLogger("filesystem.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}
}

func main() {
	logger.Info("--- Comienzo ejecución del filesystem ---")

	filesystemPort := "8084"
	http.HandleFunc("/filesystem/doSomething", doSomething)
	http.HandleFunc("/", notFound)

	logger.Info("Corriendo filesystem en el puerto %v", filesystemPort)
	log.Fatal(http.ListenAndServe("localhost:"+filesystemPort, nil))

}

func doSomething(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request recibida: %v, desde %v", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Hola recibi la request"))
	if err != nil {
		logger.Error("No se pudo escribir la respuesta - %v", err.Error())
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request inválida %v", r.RequestURI)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request inválida"))
	if err != nil {
		logger.Error("No se pudo escribir la respuesta - %v", err.Error())
	}
}
