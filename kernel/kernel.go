package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var Config KernelConfig

func init() {
	// Init logger
	err := logger.ConfigureLogger("kernel.log", "INFO")
	if err != nil {
		fmt.Println("No se pudo crear el logger -", err.Error())
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load Config
	configData, err := os.ReadFile("Config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &Config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	if err = Config.validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

}

func main() {
	logger.Info("-- Comenzó la ejecución del kernel --")

	// Implementacion GOD
	// Inicializar un proceso sin que CPU mande nada
	// PROCESS_CREATE(// fileName, processSize, TID)

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	http.HandleFunc("/", badRequest)
	http.HandleFunc("POST kernel/process/finished", processFinish)

	url := fmt.Sprintf("%s:%d", Config.SelfAddress, Config.SelfPort)
	logger.Info("Server activo en %s", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

	// que quede corriendo con un hilo
	// planificadorLargoPlazo
	go planificadorCortoPlazo()
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

	// map a la libreria de syscalls
	err = ExecuteSyscall(request.Description, request.Arguments)
	if err != nil {
		logger.Fatal("Error al ejecutar la syscall: &t - %v", request.Description, err)
	}
	// planificadorLargoPlazo(request)

	w.WriteHeader(http.StatusOK)
}

func processFinish(w http.ResponseWriter, r *http.Request) {
	// Cosa de largo plazo :)
	// TODO: Hacer lo que tenga que hacer cuando termine un proceso YYY Hacer el Unlock del mutex cpu dentro de largo plazo no aca
}
