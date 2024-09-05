package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
)

var config CpuConfig
var executionContext types.ExecutionContext
var currentThread types.Thread

const (
	ProcessExitOutcome = iota
	ThreadExitOutcome
	SyscallOutcome
	PreemptionOutcome
	SegfaultOutcome
)

func init() {
	// Configure logger
	err := logger.ConfigureLogger("cpu.log", config.LogLevel)
	if err != nil {
		fmt.Printf("No se pudo crear el logger - %v\n", err)
		os.Exit(1)
	}

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

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo setear el nivel de log - %v", err.Error())
	}

}

func main() {
	logger.Info("--- Comienzo ejecución CPU ---")

	http.HandleFunc("POST /cpu/execute", executeThread)
	http.HandleFunc("/", BadRequest)
	logger.Info("CPU escuchando en puerto 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("Listen and serve retornó error - " + err.Error())
	}
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request malformada")
	w.WriteHeader(http.StatusBadRequest)
	jsonError, err := json.MarshalIndent("Bad Request", "", "  ")
	_, err = w.Write(jsonError)
	if err != nil {
		logger.Error("Error al escribir la respuesta a BadRequest")
	}
}

func executeThread(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request recibida de: %v", r.RemoteAddr)
	body, err := io.ReadAll(r.Body)

	if err != nil {
		badRequest(w, r)
		return
	}

	var execMsg types.Thread
	err = json.Unmarshal(body, &execMsg)
	if err != nil {
		badRequest(w, r)
		return
	}

	executionContext, err = memoryGiveMeExecutionContext(execMsg)
	if err != nil {
		logger.Error("No se pudo obtener el contexto de ejecución del T%v P%v - %v",
			execMsg.Tid, execMsg.Pid, err.Error())
	}

	logger.Debug("Iniciando la ejecución del hilo %v del proceso %v", execMsg.Tid, execMsg.Pid)
	outcome := loopInstructionCycle()
	if outcome < ProcessExitOutcome || outcome >= SegfaultOutcome {
		logger.Fatal("Un ciclo de instrucción retornó un outcome no posible!")
	}

	err = kernelYourProcessFinished(execMsg, outcome)
	if err != nil {
		// Yo creo que esto es suficientemente grave como para terminar la ejecución
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}

}

func loopInstructionCycle() int {
	for {
		// Fetch
		instruction, err := memoryGiveMeInstruction(currentThread, executionContext.pc)
		if err != nil {
			logger.Fatal("No se pudo obtener instrucción a ejecutar - %v", err.Error())
		}

		// Decode
		// todo: magia?

		// Execute
		// Todo: como parseamos la instrucción??
		logger.Info("T%v P%v - Ejecutando: '%v'",
			currentThread.Tid, currentThread.Pid, instruction)

		// Checkinterrupt
		// TODO: Qué son las interrupt?? quién las hace? son enums, structs, ints?

	}

}

func badRequest(w http.ResponseWriter, r *http.Request) {
	logger.Error("CPU recibió una request mal formada de %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Tu request está mal formada!"))
	if err != nil {
		logger.Error("Error escribiendo response a %v", r.RemoteAddr)
	}
}
