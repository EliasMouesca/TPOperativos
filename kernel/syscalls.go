package main

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

type syscallFunction func(args []string) error

// TODO: Dónde ponemos esto? en qué carpeta?

var syscallSet = map[int]syscallFunction{
	syscalls.ProcessCreate: ProcessCreate,
	syscalls.ProcessExit:   ProcessExit,
	syscalls.ThreadCreate:  ThreadCreate,
	// "THREAD_JOIN": THREAD_JOIN,
	// "THREAD_CANCEL": THREAD_CANCEL
	// "THREAD_EXIT": THREAD_CREATE,
	// "MUTEX_CREATE": MUTEX_CREATE,
	// "MUTEX_LOCK": MUTEX_LOCK,
	// "MUTEX_UNLOCK": MUTEX_UNLOCK,
}

func ExecuteSyscall(syscall syscalls.Syscall) error {
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	logger.Info("## (%v:%v) - Solicitó syscall: <%v>",
		kernelglobals.ExecStateThread.ConectPCB.PID,
		kernelglobals.ExecStateThread.TID,
		syscalls.SyscallNames[syscall.Type],
	)

	err := syscallFunc(syscall.Arguments)
	if err != nil {
		return err
	}

	return nil
}

var PIDcount int = 0

func ProcessCreate(args []string) error {
	// Esta syscall recibirá 3 parámetros de la CPU, el primero será el nombre del archivo
	// de pseudocódigo que deberá ejecutar el proceso, el segundo parámetro es el tamaño del proceso en
	// Memoria y el tercer parámetro es la prioridad del hilo main (TID 0). El Kernel creará un nuevo PCB y
	// un TCB asociado con TID 0 y lo dejará en estado NEW.

	// Se crea el PCB
	var processCreate kerneltypes.PCB
	PIDcount++
	processCreate.PID = PIDcount
	processCreate.TIDs = []int{0}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", processCreate.PID)

	// Se agrega el proceso a NEW
	kernelglobals.NewStateQueue.Add(&processCreate)

	kernelsync.ChannelProcessArguments <- args

	return nil
}

func ProcessExit(args []string) error {
	// Esta syscall finalizará el PCB correspondiente al TCB que ejecutó la instrucción,
	// enviando todos sus TCBs asociados a la cola de EXIT. Esta instrucción sólo será llamada por el TID 0
	// del proceso y le deberá indicar a la memoria la finalización de dicho proceso.

	tcb := kernelglobals.ExecStateThread
	pcb := tcb.ConectPCB
	queueSize := kernelglobals.ReadyStateQueue.Size()

	if tcb.TID == 0 { // tiene que ser el hiloMain
		kernelsync.ChannelFinishprocess <- pcb.PID
		<-kernelsync.SemFinishprocess

		for i := 0; i < queueSize; i++ {
			readyTCB, err := kernelglobals.ReadyStateQueue.GetAndRemoveNext()
			if err != nil {
				logger.Error("Error al obtener el siguiente TCB de ReadyStateQueue - %v", err)
			}

			// Verificar si el TCB pertenece al mismo PCB que el proceso que está finalizando
			if readyTCB.ConectPCB == pcb {
				// Si el TCB pertenece al PCB, lo eliminamos de la cola y no lo reinsertamos
				logger.Info("Eliminando TCB con TID %d del proceso con PID %d de ReadyStateQueue", readyTCB.TID, pcb.PID)
			} else {
				// Si no pertenece, lo volvemos a insertar en la cola
				kernelglobals.ReadyStateQueue.Add(readyTCB)
			}
		}
		logger.Info("## Finaliza el proceso <%v>", pcb.PID)
		// enviar señal para intentar inicializar
		// un proceso en ready
		kernelsync.InitProcess.Add(1)
	} else {
		return errors.New("El hilo que quizo eliminar el proceso, no es el hilo main")
	}

	return nil
}

func ThreadCreate(args []string) error {

	// Esta syscall recibirá como parámetro de la CPU el nombre del archivo de
	// pseudocódigo que deberá ejecutar el hilo a crear y su prioridad. Al momento de crear el nuevo hilo,
	// deberá generar el nuevo TCB con un TID autoincremental y poner al mismo en el estado READY.

	//fileName := args[0]
	prioridad, _ := strconv.Atoi(args[1])

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.ConectPCB

	newTID := len(currentPCB.TIDs)

	newTCB := kerneltypes.TCB{
		TID:       newTID,
		Prioridad: prioridad,
		ConectPCB: currentPCB,
	}

	currentPCB.TIDs = append(currentPCB.TIDs, newTID)
	kernelglobals.ReadyStateQueue.Add(&newTCB)
	logger.Info("## (<%d>:<%d>) Se crea un nuevo hilo - Estado: READY", currentPCB.PID, newTCB.TID)

	return nil
}

func THREAD_JOIN(args []string) error {
	// Esta syscall recibe como parámetro un TID, mueve el hilo que la invocó al estado
	// BLOCK hasta que el TID pasado por parámetro finalice. En caso de que el TID pasado por parámetro
	// no exista o ya haya finalizado, esta syscall no hace nada y el hilo que la invocó continuará su
	// ejecución.
	tidAFinalizar, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("error al convertir el TID a entero")
	}
	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.ConectPCB

	tidExiste := false
	for _, tid := range currentPCB.TIDs {
		if tid == tidAFinalizar {
			tidExiste = true
			break
		}
	}

	if !tidExiste {
		logger.Info("## (<%d>:<%d>) TID <%d> no existe o ya ha finalizado, continúa la ejecución. ", currentPCB.PID, execTCB.TID, tidAFinalizar)
		return nil
	}

	logger.Info("## (<%d>:<%d>) Hilo se mueve a estado BLOCK esperando a TID <%d>", currentPCB.PID, execTCB.TID, tidAFinalizar)
	kernelglobals.BlockedStateQueue.Add(&execTCB)

	logger.Info("## (<%d>:<%d>) TID <%d> se mueve a estado EXEC", currentPCB.PID, tidAFinalizar, tidAFinalizar)
	//kernelglobals.ExecStateThread =   que buscar en la cola de ready el tcb con tid = tidAFinalizar y asignarlo al execStateThread

	//var tcpAFinalizar kerneltypes.TCB
	//for _, tid := range currentPCB.TIDs {
	//	if tid == tidAFinalizar {
	//
	//		break
	//	}
	//}

	return nil
}

func THREAD_CANCEL(args []string) {
	fmt.Println("Cancelando hilo...")
}

func THREAD_EXIT(args []string) {
	fmt.Println("Saliendo del hilo...")
}

func MUTEX_CREATE(args []string) {
	fmt.Println("Creando mutex...")
}

func MUTEX_LOCK(args []string) {
	fmt.Println("Bloqueando mutex...")
}

func MUTEX_UNLOCK(args []string) {
	fmt.Println("Desbloqueando mutex...")
}
