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
	"sync"
)

type syscallFunction func(args []string) error

// TODO: Dónde ponemos esto? en qué carpeta?

var syscallSet = map[int]syscallFunction{
	syscalls.ProcessCreate: ProcessCreate,
	syscalls.ProcessExit:   ProcessExit,
	syscalls.ThreadCreate:  ThreadCreate,
	syscalls.ThreadJoin:    ThreadJoin,
	syscalls.ThreadCancel:  ThreadCancel,
	syscalls.ThreadExit:    ThreadExit,
	syscalls.MutexCreate:   MutexCreate,
	syscalls.MutexLock:     MutexLock,
	syscalls.MutexUnlock:   MutexUnlock,
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
		kernelsync.InitProcess <- 0
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

func ThreadJoin(args []string) error {
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

	var tcbAFinalizar *kerneltypes.TCB
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tidAFinalizar && tcb.ConectPCB == currentPCB {
			tcbAFinalizar = tcb
		}
	})

	if tcbAFinalizar == nil {
		return errors.New(fmt.Sprintf("no se encontró el TID <%d> en la cola de ReadyState para el PCB con PID <%d>", tidAFinalizar, currentPCB.PID))
	}

	// Remover el TCB encontrado de la ReadyStateQueue
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb == tcbAFinalizar {
			kernelglobals.ReadyStateQueue.Remove(tcb)
		}
	})

	logger.Info("## (<%d>:<%d>) TID <%d> se mueve a estado EXEC", currentPCB.PID, tcbAFinalizar.TID, tidAFinalizar)
	kernelglobals.ExecStateThread = *tcbAFinalizar

	return nil
}

func ThreadCancel(args []string) error {
	// Esta syscall recibe como parámetro un TID con el objetivo de finalizarlo
	// pasando al mismo al estado EXIT. Se deberá indicar a la Memoria la
	// finalización de dicho hilo. En caso de que el TID pasado por parámetro no
	// exista o ya haya finalizado, esta syscall no hace nada. Finalmente, el hilo
	// que la invocó continuará su ejecución.

	tidCancelar, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("error al convertir el TID a entero")
	}

	currentPCB := kernelglobals.ExecStateThread.ConectPCB

	var tcbCancelar *kerneltypes.TCB

	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tidCancelar && tcb.ConectPCB == currentPCB {
			tcbCancelar = tcb
		}
	})

	if tcbCancelar == nil {
		kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
			if tcb.TID == tidCancelar && tcb.ConectPCB == currentPCB {
				tcbCancelar = tcb
			}
		})
	}

	if tcbCancelar == nil {
		logger.Info("## No se encontró el TID <%d> en ninguna cola para el PCB con PID <%d>. Continúa la ejecución normal.", tidCancelar, currentPCB.PID)
		return nil
	}

	logger.Info("## Finalizando el TID <%d> del PCB con PID <%d>", tcbCancelar.TID, currentPCB.PID)
	// falta la logica para avisarle a memoria que se finalizo el hilo

	logger.Info("## Moviendo el TID <%d> al estado EXIT", tcbCancelar.TID)
	kernelglobals.ExitStateQueue.Add(tcbCancelar)

	kernelglobals.ReadyStateQueue.Remove(tcbCancelar)
	kernelglobals.BlockedStateQueue.Remove(tcbCancelar)

	return nil
}

func ThreadExit(args []string) error {
	// Esta syscall finaliza al hilo que la invocó, pasando el mismo al estado EXIT.
	// Se deberá indicar a la Memoria la finalización de dicho hilo.

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.ConectPCB

	logger.Info("## Finalizando el TID <%d> del PCB con PID <%d>", execTCB.TID, currentPCB.PID)
	// falta la logica para avisarle a memoria que se finalizo el hilo

	logger.Info("## Moviendo el TID <%d> al estado EXIT", execTCB.TID)
	kernelglobals.ExitStateQueue.Add(&execTCB)

	kernelglobals.ExecStateThread = kerneltypes.TCB{
		TID:       -1,
		ConectPCB: currentPCB,
	}

	return nil
}

func MutexCreate(args []string) error {
	// Crea un nuevo mutex para el proceso sin asignar a ningún hilo.

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.ConectPCB
	newMutexID := len(kernelglobals.GlobalMutexRegistry) + 1

	newMutexWrapper := &kerneltypes.MutexWrapper{
		Mutex:          sync.Mutex{},
		ID:             newMutexID,
		AssignedTID:    -1,
		BlockedThreads: []*kerneltypes.TCB{},
	}

	kernelglobals.GlobalMutexRegistry[newMutexID] = newMutexWrapper

	currentPCB.Mutex = append(currentPCB.Mutex, newMutexID)

	logger.Info("## Se creó el mutex <%d> para el proceso con PID <%d>", newMutexID, currentPCB.PID)

	return nil
}

func MutexLock(args []string) error {

	mutexID, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("error al convertir el ID del mutex a entero")
	}

	execTCB := kernelglobals.ExecStateThread

	mutexWrapper, exists := kernelglobals.GlobalMutexRegistry[mutexID]
	if !exists {
		return errors.New(fmt.Sprintf("No se encontró el mutex con ID <%d>", mutexID))
	}

	mutexWrapper.Mutex.Lock()
	defer mutexWrapper.Mutex.Unlock()

	if mutexWrapper.AssignedTID != -1 && mutexWrapper.AssignedTID != execTCB.TID {
		mutexWrapper.BlockedThreads = append(mutexWrapper.BlockedThreads, &execTCB)
		logger.Info("## El mutex <%d> ya está tomado. Bloqueando al TID <%d> del proceso con PID <%d>", mutexID, execTCB.TID, execTCB.ConectPCB.PID)
		return nil
	}

	mutexWrapper.AssignedTID = execTCB.TID
	execTCB.Mutex = append(execTCB.Mutex, mutexID)
	logger.Info("## El mutex <%d> ha sido asignado al TID <%d> del proceso con PID <%d>", mutexID, execTCB.TID, execTCB.ConectPCB.PID)
	kernelglobals.ExecStateThread = execTCB

	return nil
}

func MutexUnlock(args []string) error {
	// Se deberá verificar primero que exista el mutex solicitado y esté
	// tomado por el hilo que realizó la syscall. En caso de que
	// corresponda, se deberá desbloquear al primer hilo de la cola de
	// bloqueados de ese mutex y le asignará el mutex al hilo recién
	// desbloqueado. Una vez hecho esto, se devuelve la ejecución al hilo
	// que realizó la syscall MUTEX_UNLOCK. En caso de que el hilo que
	// realiza la syscall no tenga asignado el mutex, no realizará
	// ningún desbloqueo.

	mutexID, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("error al convertir el ID del mutex a entero")
	}

	execTCB := kernelglobals.ExecStateThread

	mutexWrapper, exists := kernelglobals.GlobalMutexRegistry[mutexID]
	if !exists {
		return errors.New(fmt.Sprintf("No se encontró el mutex con ID <%d>", mutexID))
	}

	mutexWrapper.Mutex.Lock()
	defer mutexWrapper.Mutex.Unlock()

	if mutexWrapper.AssignedTID != execTCB.TID {
		logger.Info("## El hilo actual (TID <%d>) no tiene asignado el mutex <%d>. No se realizará ningún desbloqueo.", execTCB.TID, mutexID)
		return nil
	}

	mutexWrapper.AssignedTID = -1

	var newMutexList []int
	for _, tcbMutexID := range execTCB.Mutex {
		if tcbMutexID != mutexID {
			newMutexList = append(newMutexList, tcbMutexID)
		}
	}
	execTCB.Mutex = newMutexList

	if len(mutexWrapper.BlockedThreads) > 0 {
		nextThread := mutexWrapper.BlockedThreads[0]
		mutexWrapper.BlockedThreads = mutexWrapper.BlockedThreads[1:]

		// Asignar el mutex al hilo desbloqueado
		nextThread.Mutex = append(nextThread.Mutex, mutexID)
		mutexWrapper.AssignedTID = nextThread.TID
		logger.Info("## El mutex <%d> ha sido reasignado al TID <%d> del proceso con PID <%d>", mutexID, nextThread.TID, nextThread.ConectPCB.PID)

		kernelglobals.ReadyStateQueue.Add(nextThread)
		kernelglobals.ExecStateThread = *nextThread

	} else {
		logger.Info("## No hay hilos bloqueados esperando el mutex <%d>. Se ha liberado.", mutexID)
		kernelglobals.ExecStateThread = kerneltypes.TCB{}
	}

	return nil
}
