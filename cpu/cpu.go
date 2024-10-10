package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// TODO: Testear _algo_ lel
// TODO: Terminar los helpers

// TODO: Qué tan mal está usar globales?
// Configuración general de la CPU
var config CpuConfig

// Execution context actual, los registros que está fisicamente en la CPU
var currentExecutionContext types.ExecutionContext

// El hilo (PID + TID) que se está ejecutando en este momento
var currentThread *types.Thread = nil

// Si a este canal se le pasa una interrupción, la CPU se detiene y llama al kernel pasándole la interrupción que se haya cargado
var interruptionChannel = make(chan types.Interruption, 1)

// El momento en que se detecta la syscall es distinto del momento en que se la manda al kernel, por eso tenemos un buffer
var syscallBuffer *syscalls.Syscall

// Un mutex para la CPU porque se hay partes del código que asumen que la CPU es única por eso tenemos que excluir mutuamente
// las distintas requests que llegen (aunque el kernel en realidad nunca debería mandar a ejecutar un segundo hilo si
// el primero no terminó, pero bueno, por las dudas.
var cpuMutex = sync.Mutex{}

func init() {
	// Tatrá de configurar el logger con un nivel arbitrario
	err := logger.ConfigureLogger("cpu.log", "ERROR")
	if err != nil {
		// Si no podemos logear, no corremos
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

	// Seteamos el nivel que realmente dice la config
	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo setear el nivel de log - %v", err.Error())
	}
}

func main() {
	dino.Dino(true)
	logger.Info("--- Comienzo ejecución CPU ---")

	// Router
	http.HandleFunc("POST /cpu/interrupt", interruptFromKernel)
	http.HandleFunc("POST /cpu/execute", executeThread)
	http.HandleFunc("/", badRequest)

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Info("CPU Sirviendo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("Listen and serve retornó error - " + err.Error())
	}
}

func interruptFromKernel(w http.ResponseWriter, r *http.Request) {
	// Log request
	logger.Debug("Request %v - %v %v", r.RemoteAddr, r.Method, r.URL)

	// Parse body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("Error leyendo el body")
		badRequest(w, r)
		return
	}

	// Parse interruption
	var interruption = types.Interruption{}
	err = json.Unmarshal(body, &interruption)
	if err != nil {
		logger.Error("Error parseando la interrupción recibida en body")
		badRequest(w, r)
		return
	}

	if currentThread == nil {
		logger.Debug("No hay nada para interrumpir! Saliendo...")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Kernel, mi amor, todavía no me mandaste a ejecutar nada, qué querés que interrumpa???"))
		return
	}

	logger.Info("## Interrupción externa recibida parseada correctamente: '%v'", interruption.Description)
	if len(interruptionChannel) == 0 {
		logger.Debug("Enviando interrupción por el canal de interrupciones")
		interruptionChannel <- interruption
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("La CPU recibió la interrupción"))
	} else {
		logger.Debug("Ya se dio otra interrupción previamente, ignorando...")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("La CPU ya recibió otra interrupción y se va a detener al final del ciclo"))
	}
}

func executeThread(w http.ResponseWriter, r *http.Request) {
	// Log request
	logger.Debug("Request de %v - %v", r.RemoteAddr, r.URL)

	query := r.URL.Query()
	tid, err := strconv.Atoi(query.Get("tid"))
	if err != nil {
		logger.Error("Query param no se pudo traducir a int - %v", err.Error())
		badRequest(w, r)
		return
	}

	pid, err := strconv.Atoi(query.Get("pid"))
	if err != nil {
		logger.Error("Query param no se pudo traducir a int - %v", err.Error())
		badRequest(w, r)
		return
	}
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	// Esperá a que la CPU esté libre, no pinta andar cambiándole el contexto y el currentThread al proceso que se está ejecutando
	cpuMutex.Lock()

	// Obtenemos el contexto de ejecución
	logger.Info("Proceso P%v T%v admitido en la CPU", thread.PID, thread.TID)
	logger.Debug("Obteniendo contexto de ejecución")
	currentExecutionContext, err = memoryGiveMeExecutionContext(thread)
	if err != nil {
		logger.Error("No se pudo obtener el contexto de ejecución del T%v P%v - %v",
			thread.TID, thread.PID, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No se pudo obtener el contexto de ejecución - " + err.Error()))
		cpuMutex.Unlock()
		return
	}

	// Si hasta acá las cosas salieron bien, poné a ejecutar el proceso
	logger.Debug("Iniciando la ejecución del hilo %v del proceso %v", thread.TID, thread.PID)
	currentThread = &thread
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
		logger.Info("## T%v P%v - Ejecutando: '%v' %v",
			currentThread.TID, currentThread.PID, instructionToParse, arguments)

		err = instruction(&currentExecutionContext, arguments)
		if err != nil {
			logger.Error("No se pudo ejecutar la instrucción - %v", err.Error())
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

	finishedThread := *currentThread
	finishedExecutionContext := currentExecutionContext
	receivedInterrupt := <-interruptionChannel
	currentThread = nil

	// Libera la CPU
	cpuMutex.Unlock()

	// Kernel tu proceso terminó
	err := kernelYourProcessFinished(finishedThread, receivedInterrupt)
	if err != nil {
		// Yo creo que esto es suficientemente grave como para terminar la ejecución
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}

	err = memoryUpdateExecutionContext(finishedThread, finishedExecutionContext)
	if err != nil {
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}

}

func fetch() (instructionToParse string, err error) {
	instructionToParse, err = memoryGiveMeInstruction(*currentThread, currentExecutionContext.Pc)
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
	//logger.Error("CPU recibió una request mal formada de %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Tu request está mal formada!"))
	if err != nil {
		logger.Error("Error escribiendo response a %v", r.RemoteAddr)
	}
}
