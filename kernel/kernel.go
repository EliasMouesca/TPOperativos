package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
	"strconv"
	"time"
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

	if err = kernelglobals.Config.Validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}

	err = logger.SetLevel(kernelglobals.Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

	// Esto es un semaforo binario, arranca en 0, no tenog una mejor solución que mandarle el valor acá.
	go func() {
		kernelsync.SyscallFinalizada <- true
	}()
}

func main() {
	dino.Trex()
	logger.Debug("-- Comenzó la ejecución del kernel --")

	// TODO: Que onda lo del tamaño, preguntar bien si lo podemos sacar a la mierda de una vez!
	// -- PARSEAMOS LOS ARGS --
	// Capturar los argumentos pasados al kernel por consola
	/*
		if len(os.Args) < 3 {
			logger.Fatal("Se requieren al menos dos argumentos: archivo de pseudocódigo y tamaño del proceso.")
		}
	*/

	if len(os.Args) < 2 {
		logger.Fatal("Se requiere al menos un argumento, el archivo a ejecutar")
	}

	fileName := os.Args[1]    // Primer argumento: nombre del archivo de pseudocódigo
	processSize := os.Args[2] // Segundo argumento: tamanio del proceso

	//if _, err := os.Stat(fileName); os.IsNotExist(err) {
	//	logger.Fatal("El archivo '%v' no existe!", fileName)
	//}

	// -- INICIALIZAMOS KERNEL --
	// Inicializar colas y planificadores globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}

	// Crear el primer proceso
	logger.Debug("## Creando el proceso inicial (archivo: %s, tamaño: %s)", fileName, processSize)

	// Inicializamos el planificador de corto plazo (PCP)
	logger.Info("Iniciando el planificador de corto plazo: %v", kernelglobals.Config.SchedulerAlgorithm)
	kernelglobals.ShortTermScheduler = AlgorithmsMap[kernelglobals.Config.SchedulerAlgorithm]

	// Poner a correr corto plazo
	go planificadorCortoPlazo()
	initFirstProcess(fileName, processSize)

	planificadorLargoPlazo()

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	http.HandleFunc("/kernel/process_finished", CpuReturnThread)
	http.HandleFunc("/", badRequest)

	url := fmt.Sprintf("%s:%d", kernelglobals.Config.SelfAddress, kernelglobals.Config.SelfPort)
	logger.Debug("Server activo en %s", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

}

func badRequest(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request inválida: %v", r.RequestURI)
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
	err = ExecuteSyscall(syscall) // map a la libreria de syscalls
	if err != nil {
		// Por alguna razón esto rompe cuando quiero compilar
		logger.Error("Error al ejecutar la syscall: %v - %v", syscalls.SyscallNames[syscall.Type], err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	logger.Trace("SE ejecuto ExecuteSyscall()")
	logCurrentState("Estado luego de recibir syscall.")
	w.WriteHeader(http.StatusOK)
}

func ExecuteSyscall(syscall syscalls.Syscall) error {
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	logger.Debug("Espera en execute syscall a que termine la planificacion")
	<-kernelsync.PlanificacionFinalizada
	logger.Debug("Termino la planificacion, ejecutando syscall")

	// Verificar si hay un thread en ejecución
	if kernelglobals.ExecStateThread != nil {
		logger.Info("## (<%v>:<%v>) - Solicitó syscall: <%v>",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
			syscalls.SyscallNames[syscall.Type],
		)
	} else {
		logger.Error("Syscall solicitada <%v>, pero no hay un thread en ejecución actualmente", syscalls.SyscallNames[syscall.Type])
		return nil
	}
	// Si es CMN "pausamos" el quantum y decimos que hay restante
	if kernelglobals.Config.SchedulerAlgorithm == "CMN" {
		logger.Debug("Seteando ExitInstant")
		kernelglobals.ExecStateThread.ExitInstant = time.Now()
		kernelglobals.ExecStateThread.HayQuantumRestante = true
	}

	// Ejecuta la syscall
	kernelsync.MutexExecThread.Lock()
	err := syscallFunc(syscall.Arguments)
	kernelsync.MutexExecThread.Unlock()

	if kernelglobals.ExecStateThread == nil {
		logger.Trace("ExecStateThread post syscall: nil")
	} else {
		logger.Trace("ExecStateThread post syscall: (PID: %v - TID: %v)", kernelglobals.ExecStateThread.FatherPCB.PID, kernelglobals.ExecStateThread.TID)
	}

	if err != nil {
		logger.Error("La syscall devolvió un error - %v", err)
	}

	logger.Debug("-- Termino de ejecutar syscall --")
	go func() {
		logger.Debug("Mandando al channel SyscallFinalizada")
		kernelsync.SyscallFinalizada <- true

		if kernelglobals.Config.SchedulerAlgorithm == "CMN" {
			logger.Debug("Mandando a SyscallChannel")
			kernelsync.SyscallChannel <- struct{}{}
			logger.Debug("Despues de mandar a SyscallChannel")
		}
	}()

	return nil
}

func initFirstProcess(fileName, processSize string) {
	// Crear el PCB para el proceso inicial
	pid := types.Pid(0) // Asignar el primer PID como 1 (puedes cambiar según la lógica de PID en tu sistema)
	pcb := kerneltypes.PCB{
		PID:  pid,
		TIDs: []types.Tid{0}, // El primer TCB tiene TID 0
	}

	logger.Info("## (<%v>:0) Se crea el proceso - Estado: NEW", pid)

	// Agregar el PCB a la lista global de PCBs en el kernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)

	// LE AVISO A MEMORIA QUE SE CREO UN NUEVO PROCESO
	request := types.RequestToMemory{
		Thread:    types.Thread{PID: types.Pid(pid)},
		Type:      types.CreateProcess,
		Arguments: []string{fileName, processSize},
	}
	for {
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error al enviar request a memoria: %v", err)
			//<-kernelsync.InitProcess // Espera a que finalice otro proceso antes de intentar de nuevo
		} else {
			logger.Debug("Hay espacio disponible en memoria")
			break
		}
	}

	// Crear el TCB (thread) principal con TID 0 y prioridad 0
	mainThread := kerneltypes.TCB{
		TID:             0,
		Prioridad:       0,                      // Prioridad más alta (0)
		FatherPCB:       &pcb,                   // Asociar el TCB al PCB creado
		LockedMutexes:   []*kerneltypes.Mutex{}, // Sin mutex bloqueados al inicio
		JoinedTCB:       nil,                    // No está unido a ningún otro thread
		QuantumRestante: time.Duration(kernelglobals.Config.Quantum) * time.Millisecond,
	}

	// Agregar el TCB a la lista global de TCBs en el kernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread)

	//// LE AVISO A MEMORIA QUE SE CREO UN NUEVO HILO
	//request2 := types.RequestToMemory{
	//	Thread:    types.Thread{PID: mainThread.FatherPCB.PID, TID: mainThread.TID},
	//	Type:      types.CreateThread,
	//	Arguments: []string{fileName},
	//}
	//
	//// Enviar la solicitud a memoria
	//err2 := sendMemoryRequest(request2)
	//if err2 != nil {
	//	logger.Error("Error en el request a memoria: %v", err2)
	//	return
	//}

	// Hacer que este thread sea el que está en ejecución
	newThread := buscarTCBPorTID(mainThread.TID, pcb.PID)
	err := kernelglobals.ShortTermScheduler.AddToReady(newThread)
	if err != nil {
		logger.Fatal("No se pudo poner en ready el primer hilo")
	}

	logCurrentState("Estado general luego de Inicializar Kernel")
}

func CpuReturnThread(w http.ResponseWriter, r *http.Request) {

	tid, err := strconv.Atoi(r.URL.Query().Get("tid"))
	pid, err := strconv.Atoi(r.URL.Query().Get("pid"))
	thread := types.Thread{
		TID: types.Tid(tid),
		PID: types.Pid(pid),
	}
	var interruption types.Interruption
	err = json.NewDecoder(r.Body).Decode(&interruption)
	if err != nil {
		logger.Error("No se puedo leer la request de CPU")
	}

	logger.Debug("Se recibio la interrupcion < %v > de CPU", interruption.Description)

	kernelsync.MutexCPU.Unlock()
	logger.Trace("Unlockear la CPU")

	// Encontrá nuestro TCB
	for _, tcb := range kernelglobals.EveryTCBInTheKernel {
		if tcb.TID == thread.TID && tcb.FatherPCB.PID == thread.PID {
			//logger.Debug("Seteando ExitInstant")
			//tcb.ExitInstant = time.Now()

			// Si la INT fue eviction o EOQ -> vuelve a ready
			if interruption.Type == types.InterruptionEviction ||
				interruption.Type == types.InterruptionEndOfQuantum {

				logger.Debug("Se pone ExecStateThread en nil por fin de quantum o desalojo => Hay que replanificar")
				kernelglobals.ExecStateThread = nil

				logger.Info("## (<%v>:<%v>) Se agrega a cola READY despues de EndOfQuantum o Desalojo", tcb.FatherPCB.PID, tcb.TID)
				err = kernelglobals.ShortTermScheduler.AddToReady(&tcb)

				kernelsync.SyscallFinalizada <- true

			} else if interruption.Type == types.InterruptionSyscall {

			} else {
				logger.Debug("Matando hilo: %v", tcb.TID)
				// Sino, muere (ie: sysegv, div por 0, ...)
				killTcb(&tcb)
				kernelsync.SyscallChannel <- struct{}{}
				kernelglobals.ExecStateThread = nil
				kernelsync.SyscallFinalizada <- true
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	logger.Error("La cpu sacó de ejecución un proceso que el kernel no conocía -> xd")
	w.WriteHeader(http.StatusNotFound)
	return
}

func killTcb(tcb *kerneltypes.TCB) {
	pcbPadre := tcb.FatherPCB
	// Elimino el tid del pcb padre
	for i, tid := range pcbPadre.TIDs {
		if tid == tcb.TID {
			pcbPadre.TIDs = append(pcbPadre.TIDs[:i], pcbPadre.TIDs[i+1:]...)
			break
		}
	}

	// como despues voy a sacar el lockedMutex de la lista de lockedMutexes va a alterar el recorrido de la lista, entonces la recorro en orden inverso
	for i := len(tcb.LockedMutexes) - 1; i >= 0; i-- {
		lockedMutex := tcb.LockedMutexes[i]
		lockedMutex.AssignedTCB = nil

		// lo saco de la lista de mutexes lockeados por el tcb
		tcb.LockedMutexes = append(tcb.LockedMutexes[:i], tcb.LockedMutexes[i+1:]...)

		if len(lockedMutex.BlockedTCBs) > 0 {
			nextTcb := lockedMutex.BlockedTCBs[0]
			lockedMutex.BlockedTCBs = lockedMutex.BlockedTCBs[1:]

			// asigno el mutex al siguiente TCB y lo pongo en ready
			nextTcb.LockedMutexes = append(nextTcb.LockedMutexes, lockedMutex)
			lockedMutex.AssignedTCB = nextTcb

			err := kernelglobals.ShortTermScheduler.AddToReady(nextTcb)
			if err != nil {
				logger.Error("Error al mover el TCB <%d> a Ready - %v", nextTcb.TID, err)
			}
		}
	}
	kernelglobals.ExitStateQueue.Add(tcb)
	logger.Info("## (<%d:%d>) Movido a la cola Exit por respuesta de CPU", tcb.FatherPCB.PID, tcb.TID)

}
