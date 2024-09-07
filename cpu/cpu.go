package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
	"strings"
)

// TODO: Codear hilo que loopee esperando interrupciones del kernel
// TODO: Codear las instrucciones que faltan
// TODO: Testear _algo_ lel

var config CpuConfig
var executionContext types.ExecutionContext
var currentThread types.Thread
var interruptChannel = make(chan int)
var cpuIsFree = make(chan bool)

// These are the interrupts the CPU understands
const (
	ProcessExit = iota
	ThreadExit
	Syscall
	Preemption
	BadInstruction
	Segfault
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

	cpuIsFree <- true

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
	// Log request
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Parse body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		badRequest(w, r)
		return
	}

	// Parse Thread
	var execMsg types.Thread
	err = json.Unmarshal(body, &execMsg)
	if err != nil {
		badRequest(w, r)
		return
	}

	// Obtenemos el contexto de ejecución
	logger.Debug("Obteniendo contexto de ejecución")
	executionContext, err = memoryGiveMeExecutionContext(execMsg)
	if err != nil {
		logger.Error("No se pudo obtener el contexto de ejecución del T%v P%v - %v",
			execMsg.Tid, execMsg.Pid, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No se pudo obtener el contexto de ejecución - " + err.Error()))
		return
	}

	// Si hasta acá las cosas salieron bien, poné a ejecutar el proceso
	logger.Debug("Iniciando la ejecución del hilo %v del proceso %v", execMsg.Tid, execMsg.Pid)
	currentThread = execMsg
	go loopInstructionCycle()

	// Repondemos al kernel: "Tu proceso se está ejecutando, sé feliz"
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Tu proceso se está ejecutando, gut!"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err.Error())
	}
}

func loopInstructionCycle() {

	// Wait for cpu to be free
	<-cpuIsFree

	for {
		// Fetch
		instructionToParse, err := fetch()
		if err != nil {
			logger.Fatal("No se pudo obtener instrucción a ejecutar - %v", err.Error())
		}

		// Decode
		instruction, arguments, err := decode(instructionToParse)
		if err != nil {
			logger.Error("No se pudo decodificar la instrucción - %v", err.Error())

		}

		// Execute
		logger.Info("T%v P%v - Ejecutando: '%v'",
			currentThread.Tid, currentThread.Pid, instructionToParse, arguments)

		// Tanto trabajo decodificando resulta en la siguiente línea, que viva el paradigma funcional
		executionContext, err = instruction(executionContext, arguments)
		if err != nil {
			logger.Error("no se pudo ejecutar la instrucción - %v", err.Error())
			interruptChannel <- BadInstruction
		}

		// Checkinterrupt
		if len(interruptChannel) > 0 {
			break
		}

	}

	// Free up the cpu
	cpuIsFree <- true

	// Kernel tu proceso terminó, por qué? leer interruptChannel
	err := kernelYourProcessFinished(currentThread, <-interruptChannel)
	if err != nil {
		// Yo creo que esto es suficientemente grave como para terminar la ejecución
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}
}

func fetch() (instructionToParse string, err error) {
	instructionToParse, err = memoryGiveMeInstruction(currentThread, executionContext.Pc)
	if err != nil {
		return "", err
	}
	return instructionToParse, nil
}

func decode(instructionToDecode string) (instruction Instruction, arguments []string, err error) {
	instructionStringSplitted := strings.Split(instructionToDecode, " ")
	instructionString := instructionStringSplitted[0]
	arguments = instructionStringSplitted[1:]

	instruction, exists := instructionSet[instructionString]
	if !exists {
		return nil, nil, errors.New("no se conoce ninguna instrucción '" + instructionString + "'")
	}

	return instruction, arguments, nil

}

func badRequest(w http.ResponseWriter, r *http.Request) {
	logger.Error("CPU recibió una request mal formada de %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Tu request está mal formada!"))
	if err != nil {
		logger.Error("Error escribiendo response a %v", r.RemoteAddr)
	}
}
