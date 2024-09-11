package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var config KernelConfig

func init() {
	// Init logger
	err := logger.ConfigureLogger("kernel.log", "INFO")
	if err != nil {
		fmt.Println("No se pudo crear el logger -", err.Error())
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	if err = config.validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

}

func main() {
	logger.Info("-- Comenzó la ejecución del kernel --")

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	// En syscallRecieve, quiza podria ir el planificador a largo plazo como funciom
	// y otro para corto en otro handle
	http.HandleFunc("/", badRequest)

	url := fmt.Sprintf("%s:%d", config.SelfAddress, config.SelfPort)
	logger.Info("Server activo en %s", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

	// Argumentos que me tiene que pasar CPU
	// var fileName string
	// var processSize int
	// var TID int

	// PROCESS_CREATE(// fileName, processSize, TID)

}

func badRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request inválida: %v", r.RequestURI)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Bad request!"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err)
	}
}

func syscallRecieve(w http.ResponseWriter, r *http.Request) {

	var request syscalls.Syscall
	// Parsear la syscall request
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud - %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Call the long-term scheduler with the syscall name and data
	planificadorLargoPlazo(request)

	w.WriteHeader(http.StatusOK)
}
