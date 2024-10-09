package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
	"sync"
)

func init() {
	// Init logger
	err := logger.ConfigureLogger("kernel.log", "INFO")
	if err != nil {
		fmt.Println("No se pudo crear el logger -", err.Error())
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load Config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &kernelglobals.Config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	// Inicializar colas y planificadores globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{},
	}

	// Agregar logs para verificar los valores cargados
	logger.Info("Configuración cargada: SelfPort=%d, MemoryPort=%d, CpuPort=%d",
		kernelglobals.Config.SelfPort,
		kernelglobals.Config.MemoryPort,
		kernelglobals.Config.CpuPort)

	if err = kernelglobals.Config.Validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}

	err = logger.SetLevel(kernelglobals.Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

	logger.Info("Configuración cargada exitosamente")
}

func main() {
	logger.Info("-- Comenzó la ejecución del kernel --")

	//TODO: PARA INICIALIZAR EL KERNEL HAY QUE PONER EN CONSOLA:
	// go run . file_name 123

	// Capturar los argumentos pasados al kernel por consola
	if len(os.Args) < 3 {
		logger.Fatal("Se requieren al menos dos argumentos: archivo de pseudocódigo y tamaño del proceso.")
	}
	fileName := os.Args[1]    // Primer argumento: nombre del archivo de pseudocódigo
	processSize := os.Args[2] // Segundo argumento: tamaño del proceso

	// Crear el primer proceso
	logger.Info("Creando el primer proceso inicial (archivo: %s, tamaño: %s)", fileName, processSize)
	initProcess(fileName, processSize)

	go planificadorLargoPlazo()
	//go planificadorCortoPlazo()

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	http.HandleFunc("/", badRequest)

	url := fmt.Sprintf("%s:%d", kernelglobals.Config.SelfAddress, kernelglobals.Config.SelfPort)
	logger.Info("Server activo en %s", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

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

	var syscall syscalls.Syscall
	// Parsear la syscall request
	err := json.NewDecoder(r.Body).Decode(&syscall)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud - %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)

	//El WaitGroup asegura que no se envie la respuesta HTTP al cliente hasta que la syscall haya terminado

	err = ExecuteSyscall(syscall, &wg) // map a la libreria de syscalls
	if err != nil {
		// Por alguna razón esto rompe cuando quiero compilar
		logger.Error("Error al ejecutar la syscall: %v - %v", syscalls.SyscallNames[syscall.Type], err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	wg.Wait()
	logCurrentState("Estado luego de recibir syscall.")
	w.WriteHeader(http.StatusOK)
}

func ExecuteSyscall(syscall syscalls.Syscall, wg *sync.WaitGroup) error {
	defer wg.Done()
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	// Verificar si hay un thread en ejecución
	if kernelglobals.ExecStateThread != nil {
		logger.Info("## (<%v>:<%v>) - Solicitó syscall: <%v>",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
			syscalls.SyscallNames[syscall.Type],
		)
	} else {
		logger.Info("Syscall solicitada <%v>, pero no hay un thread en ejecución actualmente", syscalls.SyscallNames[syscall.Type])
	}

	go func() {
		err := syscallFunc(syscall.Arguments)
		if err != nil {
			logger.Error("La syscall devolvio un error - %v", err)
		}
	}()
	return nil
}

func initProcess(fileName, processSize string) {
	logger.Info("Inicializando el proceso inicial con archivo: %s, tamaño: %s", fileName, processSize)

	//TODO: HAY QUE CREAR ESTO A MANO POR QUE LA SYSCALL ProcessCreate NECESITA QUE HAYA UN HILO EJECUTANDO
	//		ENTONCES LO HACEMOS A MANO, QUE NO CAMBIA NADA Y SON 20 LINEAS MAS. :)

	// Crear el PCB para el proceso inicial
	pid := types.Pid(1) // Asignar el primer PID como 1 (puedes cambiar según la lógica de PID en tu sistema)
	pcb := kerneltypes.PCB{
		PID:  pid,
		TIDs: []types.Tid{0}, // El primer TCB tiene TID 0
	}

	// Agregar el PCB a la lista global de PCBs en el kernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)

	// Crear el TCB (thread) principal con TID 0 y prioridad 0
	mainThread := kerneltypes.TCB{
		TID:           0,
		Prioridad:     0,                      // Prioridad más alta (0)
		FatherPCB:     &pcb,                   // Asociar el TCB al PCB creado
		LockedMutexes: []*kerneltypes.Mutex{}, // Sin mutex bloqueados al inicio
		JoinedTCB:     nil,                    // No está unido a ningún otro thread
	}

	// Agregar el TCB a la lista global de TCBs en el kernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread)

	// Hacer que este thread sea el que está en ejecución
	newThread := &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]

	// Mover el hilo principal a la cola de ready
	kernelglobals.ExecStateThread = newThread
	logger.Info("## (<%v>:0) Se crea el proceso - Estado: NEW", pid)

	logCurrentState("Estado general luego de Inicializar Kernel")
}
