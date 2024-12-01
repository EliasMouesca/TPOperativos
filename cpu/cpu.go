package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Configuración general de la CPU
var config CpuConfig
var MutexInterruption = sync.Mutex{}

// Execution context actual, los registros que está fisicamente en la CPU
var currentExecutionContext types.ExecutionContext

// El hilo (PID + TID) que se está ejecutando en este momento
var currentThread *types.Thread = nil

// Si a este canal se le pasa una interrupción, la CPU se detiene y llama al kernel pasándole la interrupción que se haya cargado
var interruptionChannel = make(chan types.Interruption, 1)
var deudaInterrupciones = make([]types.InterrupcionInsatisfecha, 0)

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
	dino.Brachiosaurus(true)
	logger.Debug("--- Comienzo ejecución CPU ---")

	// Router
	http.HandleFunc("POST /cpu/interrupt", interruptFromKernel)
	http.HandleFunc("POST /cpu/execute", executeThread)
	http.HandleFunc("/", badRequest)

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Debug("CPU Sirviendo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("Listen and serve retornó error - " + err.Error())
	}
}

func interruptFromKernel(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Llega interrupcion de kernel")
	hiloEjecutando := currentThread

	MutexInterruption.Lock()
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

	logger.Debug("Interrupcion recibida: %v", interruption.Description)

	if hiloEjecutando == nil {
		logger.Debug("No hay nada para interrumpir! Saliendo...")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Kernel, todavía no me mandaste a ejecutar nada"))
		return
	}

	logger.Info("## Interrupción externa recibida parseada correctamente: '%v'", interruption.Description)
	if len(interruptionChannel) == 0 {
		logger.Debug("Enviando interrupción por el canal de interrupciones: %v", interruption.Description)
		interruptionChannel <- interruption
		kernelYourProcessFinished(*hiloEjecutando, interruption)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("La CPU recibió la interrupción"))
	} else {
		logger.Debug("Ya se dio otra interrupción previamente, incrementando deuda : %v", interruption.Description)
		interrupcionInsatisfecha := types.InterrupcionInsatisfecha{
			Thread:       hiloEjecutando,
			Interruption: interruption,
		}
		deudaInterrupciones = append(deudaInterrupciones, interrupcionInsatisfecha)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("La CPU ya recibió otra interrupción, incrementando deuda"))
	}
	MutexInterruption.Unlock()
}

func executeThread(w http.ResponseWriter, r *http.Request) {
	// Log request
	logger.Debug("Request de %v - %v", r.RemoteAddr, r.URL)

	var thread types.Thread
	err := json.NewDecoder(r.Body).Decode(&thread)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo del request - %v", err.Error())
		badRequest(w, r)
		return
	}

	defer r.Body.Close()

	// Esperá a que la CPU esté libre, no pinta andar cambiándole el contexto y el currentThread al proceso que se está ejecutando
	cpuMutex.Lock()
	// Obtenemos el contexto de ejecución
	logger.Debug("Proceso (<%d:%d>) admitido en la CPU", thread.PID, thread.TID)
	currentThread = &thread

	logger.Debug("Obteniendo contexto de ejecución")
	currentExecutionContext, err = memoryGiveMeExecutionContext(thread)
	if err != nil {
		logger.Error("No se pudo obtener el contexto de ejecución del T%v P%v - %v", thread.TID, thread.PID, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No se pudo obtener el contexto de ejecución - " + err.Error()))
		cpuMutex.Unlock()
		return
	}

	// Si hasta acá las cosas salieron bien, poné a ejecutar el proceso
	logger.Debug("Iniciando la ejecución del hilo %v del proceso %v", thread.TID, thread.PID)
	go loopInstructionCycle()

	// Repondemos al kernel: "Tu proceso se está ejecutando, sé feliz"
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Tu proceso se está ejecutando, gut!"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err.Error())
	}
}

func loopInstructionCycle() {
	var instructionToParse string
	var instruction Instruction
	var arguments []string
	var err error

	for {
		logger.Trace("Tomando MutexInterruption...")
		MutexInterruption.Lock()
		logger.Trace("MutexInterruption tomado")

		for i, d := range deudaInterrupciones {
			if d.Thread.Equals(currentThread) {
				logger.Warn("Tenia deuda pendiente y se va a atender: %v", d.Interruption.Description)
				interruptionChannel <- d.Interruption
				deudaInterrupciones = append(deudaInterrupciones[:i], deudaInterrupciones[i+1:]...)
				goto CheckInterrupt
			}
		}

		// Fetch
		instructionToParse, err = fetch()
		if err != nil {
			logger.Fatal("No se pudo obtener instrucción a ejecutar - %v", err.Error())
		}
		// Increment PC
		currentExecutionContext.Pc += 1

		// Decode
		instruction, arguments, err = decode(instructionToParse)
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

	CheckInterrupt:

		// Checkinterrupt
		if len(interruptionChannel) > 0 {
			interruption := <-interruptionChannel
			interruptionChannel <- interruption
			if interruption.Type == types.InterruptionEviction || interruption.Type == types.InterruptionEndOfQuantum {
				MutexInterruption.Unlock()
			}
			logger.Debug("Hay interrupcion en interruptionChannel")
			break
		} else {
			MutexInterruption.Unlock()
			logger.Debug("No hay interrupcion en interruptionChannel, continua ejecucion")
		}
	}

	MutexInterruption.Lock()
	logger.Debug("MutexInterruption tomado")
	finishedThread := *currentThread
	finishedExecutionContext := currentExecutionContext
	receivedInterrupt := <-interruptionChannel
	//currentThread = nil
	logger.Debug("La interrupcion recibida es: %v", receivedInterrupt.Description)
	// Kernel tu proceso terminó
	err = kernelYourProcessFinished(finishedThread, receivedInterrupt)
	if err != nil {
		// Yo creo que esto es suficientemente grave como para terminar la ejecución
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}

	err = memoryUpdateExecutionContext(finishedThread, finishedExecutionContext)
	if err != nil {
		logger.Fatal("No se pudo avisar al kernel de la finalización del proceso - %v", err.Error())
	}

	// Libera la CPU
	cpuMutex.Unlock()
	MutexInterruption.Unlock()
}

func fetch() (instructionToParse string, err error) {
	instructionToParse, err = memoryGiveMeInstruction(*currentThread, currentExecutionContext.Pc)
	if err != nil {
		logger.Error("Error al obtener instrucción (PID: %v, TID: %v, PC: %v): %v", currentThread.PID, currentThread.TID, currentExecutionContext.Pc, err)
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
