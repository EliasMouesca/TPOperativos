package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"log"
	"net/http"
)

type BodyRequest struct {
	Message string `json:"message"`
	Origin  string `json:"origin"`
}

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("filesystem.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}
}

func main() {
	logger.Info("--- Comienzo ejecución del filesystem ---")

	generateRequest("memoria", "8082")

	filesystemPort := "8083"
	http.HandleFunc("/filesystem/doSomething", doSomething)
	http.HandleFunc("/", notFound)

	logger.Info("Corriendo filesystem en el puerto %v", filesystemPort)
	log.Fatal(http.ListenAndServe("localhost:"+filesystemPort, nil))

}

func generateRequest(receiver string, port string) {
	// Defino la structura que acepta memoria
	receiverStruct := BodyRequest{
		Message: "Hola " + receiver,
		Origin:  "Filesystem",
	}

	// Serializo la estructura de memoria a JSON
	receiverjson, err := json.Marshal(receiverStruct)
	if err != nil {
		logger.Error("Error al serializar json - %v", err.Error())
	}
	// Convertir los bytes JSON a un io.Reader
	receiverRequest := bytes.NewBuffer(receiverjson)
	// POST
	receiverResponse, err := http.Post("http://localhost:"+port+"/memoria/accion", "application/json", receiverRequest)
	if err != nil {
		logger.Error("Error al conectar con memoria - %v", err.Error())
	} else {
		logger.Info("Conección con memoria sastifactoria - %v", receiverResponse.StatusCode)
	}
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
	logger.Debug("Request inválida %v, desde %v", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request inválida"))
	if err != nil {
		logger.Error("No se pudo escribir la respuesta - %v", err.Error())
	}
}
