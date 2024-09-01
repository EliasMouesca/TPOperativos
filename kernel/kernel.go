package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func init() {
	loggerLevel := logger.LevelInfo
	err := logger.ConfigureLogger("kernel.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}
}

func main() {
	logger.Info("-- Comenzó la ejecución del kernel --")

	// Probar conexiones con otros módulos
	// cpu
	{
		cpuPort := "8080"
		data := struct {
			Message string `json:"message"`
			Origin  string `json:"origin"`
		}{
			Message: "Hola cpu this is kernel speaking !!",
			Origin:  "Kernel",
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			logger.Error("Error al serializar json - ", err)
		}

		cpuResponse, err := http.Post("http://localhost:"+cpuPort+"/cpu/accion", "application/json",
			bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Error("No se obtuvo respuesta de la cpu! - %v", err)
		} else {
			logger.Info("Hola cpu! status code: %v", cpuResponse.StatusCode)
		}
	}

	// Listen and serve
	hostname := "localhost"
	port := "8081"

	http.HandleFunc("POST /kernel/accion", ActionDemo)
	http.HandleFunc("/", NotFound)

	logger.Info("Server activo en %v:%v", hostname, port)
	err := http.ListenAndServe(hostname+":"+port, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

}

func NotFound(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request inválida: %v", r.RequestURI)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Bad request!"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err)
	}
}

func ActionDemo(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request 'accion': %v", r.RequestURI)
	w.WriteHeader(http.StatusOK)
}
