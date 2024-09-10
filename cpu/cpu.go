package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
	"strings"
)

// TODO: Testear _algo_ lel
// TODO: Terminar los helpers

// TODO: Qué tan mal está usar globales?
// Configuración general de la CPU
var config CpuConfig

// Execution context actual, los registros que está fisicamente en la CPU
var currentExecutionContext types.ExecutionContext

// El hilo (PID + TID) que se está ejecutando en este momento
var currentThread types.Thread

// Si a este canal se le pasa una interrupción, la CPU se detiene y llama al kernel pasándole la interrupción que se haya cargado
var interruptionChannel = make(chan types.Interruption, 1)

// El momento en que se detecta la syscall es distinto del momento en que se la manda al kernel, por eso tenemos un buffer
var syscallBuffer *syscalls.Syscall

// Un mutex para la CPU porque se hay partes del código que asumen que la CPU es única por eso tenemos que excluir mutuamente
// las distintas requests que llegen (aunque el kernel en realidad nunca debería mandar a ejecutar un segundo hilo si
// el primero no terminó, pero bueno, por las dudas.
// TODO: Está bien usar un canal como mutex?
var cpuMutex = make(chan bool)

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

	cpuMutex <- true

	http.HandleFunc("POST /cpu/interrupt", interruptFromKernel)
	http.HandleFunc("POST /cpu/execute", executeThread)
	http.HandleFunc("/", BadRequest)
	logger.Info("CPU escuchando en puerto 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("Listen and serve retornó error - " + err.Error())
	}
}

func interruptFromKernel(w http.ResponseWriter, r *http.Request) {
	// Log request
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Parse body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		badRequest(w, r)
		return
	}

	// Parse interruption
	var interruption = types.Interruption{}
	err = json.Unmarshal(body, &interruption)
	if err != nil {
		badRequest(w, r)
		return
	}

	logger.Debug("Interrupción externa recibida %v", interruption.Description)
	if len(interruption.Description) == 0 {
		interruptionChannel <- interruption
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("La CPU recibió la interrupción"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("La CPU ya recibió otra interrupción y se va a detener al final del ciclo"))
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

	// Esperá a que la CPU esté libre, no pinta andar cambiándole el contexto y el currentThread al proceso que se está ejecutando
	<-cpuMutex

	// Obtenemos el contexto de ejecución
	logger.Debug("Proceso P%v T%v admitido en la CPU", execMsg.Pid, execMsg.Tid)
	logger.Debug("Obteniendo contexto de ejecución")
	currentExecutionContext, err = memoryGiveMeExecutionContext(execMsg)
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

		err = instruction(&currentExecutionContext, arguments)
		if err != nil {
			logger.Error("no se pudo ejecutar la instrucción - %v", err.Error())
			if len(interruptionChannel) != 0 {
				interruptionChannel <- types.Interruption{
					Type:        types.InterruptionBadInstruction,
					Description: "La CPU recibió una instrucción no reconocida",
				}
			}
		}

		// Checkinterrupt
		if len(interruptionChannel) > 0 {
			break
		}

	}

	finishedThread := currentThread
	receivedInterrupt := <-interruptionChannel

	// Libera la CPU
	cpuMutex <- true

	// Kernel tu proceso terminó
	err := kernelYourProcessFinished(finishedThread, receivedInterrupt)
	if err != nil {
		// Yo creo que esto es suficientemente grave como para terminar la ejecución
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}
}

func fetch() (instructionToParse string, err error) {
	instructionToParse, err = memoryGiveMeInstruction(currentThread, currentExecutionContext.Pc)
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
