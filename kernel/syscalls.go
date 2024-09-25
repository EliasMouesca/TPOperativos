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

	if tcb.TID != 0 {
		return errors.New("el hilo que quiso eliminar el proceso no es el hilo main")
	}

	kernelsync.ChannelFinishprocess <- pcb.PID
	<-kernelsync.SemFinishprocess

	// Eliminar todos los hilos del PCB de las colas de Ready
	for _, tid := range pcb.TIDs {
		existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(tid, pcb.PID)
		if existsInReady {
			err := kernelglobals.ShortTermScheduler.ThreadRemove(tid, pcb.PID)
			if err != nil {
				logger.Error("Error al eliminar el TID <%d> del PCB con PID <%d> de las colas de Ready - %v", tid, pcb.PID, err)
			} else {
				logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de las colas de Ready y se movió a ExitStateQueue", tid, pcb.PID)
			}
		}

		// 2. Verificar y eliminar hilos en la cola de Blocked
		for !kernelglobals.BlockedStateQueue.IsEmpty() {
			blockedTCB, err := kernelglobals.BlockedStateQueue.GetAndRemoveNext()
			if err != nil {
				logger.Error("Error al obtener el siguiente TCB de BlockedStateQueue - %v", err)
				break
			}
			// Si es del PCB que se está finalizando, se mueve a ExitStateQueue
			if blockedTCB.TID == tid && blockedTCB.ConectPCB == pcb {
				kernelglobals.ExitStateQueue.Add(blockedTCB)
				logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de BlockedStateQueue y se movió a ExitStateQueue", tid, pcb.PID)
			} else {
				// Si no es, se vuelve a insertar en la cola de bloqueados
				kernelglobals.BlockedStateQueue.Add(blockedTCB)
			}
		}
	}
	logger.Info("## Finaliza el proceso <%v>", pcb.PID)
	kernelsync.InitProcess <- 0

	return nil
}

func ThreadCreate(args []string) error {

	// Esta syscall recibirá como parámetro de la CPU el nombre del archivo de
	// pseudocódigo que deberá ejecutar el hilo a crear y su prioridad. Al momento de crear el nuevo hilo,
	// deberá generar el nuevo TCB con un TID autoincremental y poner al mismo en el estado READY.

	//fileName := args[0]
	prioridad, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("error al convertir la prioridad a entero: %v", err)
	}

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.ConectPCB

	newTID := len(currentPCB.TIDs)

	newTCB := kerneltypes.TCB{
		TID:       newTID,
		Prioridad: prioridad,
		ConectPCB: currentPCB,
	}

	currentPCB.TIDs = append(currentPCB.TIDs, newTID)
	err = kernelglobals.ShortTermScheduler.AddToReady(&newTCB)
	if err != nil {
		return fmt.Errorf("error al agregar el TCB a la cola de Ready: %v", err)
	}
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

	// Verificar si el tidAFinalizar ya está en ExitStateQueue (ya ha finalizado)
	finalizado := false
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tidAFinalizar && tcb.ConectPCB == currentPCB {
			finalizado = true
		}
	})

	if finalizado {
		logger.Info("## (<%d>:<%d>) TID <%d> ya ha finalizado. Continúa la ejecución. ", currentPCB.PID, execTCB.TID, tidAFinalizar)
		return nil
	}

	// Verificar si el tidAFinalizar está en la lista de TIDs del PCB actual
	tidExiste := false
	for _, tid := range currentPCB.TIDs {
		if tid == tidAFinalizar {
			tidExiste = true
			break
		}
	}

	if !tidExiste {
		logger.Info("## (<%d>:<%d>) TID <%d> no existe. Continúa la ejecución. ", currentPCB.PID, execTCB.TID, tidAFinalizar)
		return nil
	}

	// Buscar el tcbAFinalizar en la ReadyStateQueue
	var tcbAFinalizar *kerneltypes.TCB
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tidAFinalizar && tcb.ConectPCB == currentPCB {
			tcbAFinalizar = tcb
		}
	})

	// Si el tcbAFinalizar no está en ReadyStateQueue, pero tampoco en Exit, significa que no está ejecutándose y puede causar deadlock
	if tcbAFinalizar == nil {
		return errors.New(fmt.Sprintf("No se encontró el TID <%d> en la cola de ReadyState para el PCB con PID <%d>", tidAFinalizar, currentPCB.PID))
	}

	execTCB.WaitingForTID = tidAFinalizar // Aquí se asigna el TID que está esperando.
	kernelglobals.ExecStateThread = execTCB

	logger.Info("## (<%d>:<%d>) Hilo se mueve a estado BLOCK esperando a TID <%d>", currentPCB.PID, execTCB.TID, tidAFinalizar)
	kernelglobals.BlockedStateQueue.Add(&execTCB)

	// Cambiar el hilo tcbAFinalizar a EXEC
	kernelglobals.ReadyStateQueue.Remove(tcbAFinalizar)
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

	var estaEnReady = false
	var estaEnBlocked = false

	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tidCancelar && tcb.ConectPCB == currentPCB {
			tcbCancelar = tcb
			estaEnReady = true
		}
	})

	if tcbCancelar == nil {
		kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
			if tcb.TID == tidCancelar && tcb.ConectPCB == currentPCB {
				tcbCancelar = tcb
				estaEnBlocked = true
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
	if estaEnReady {
		err = kernelglobals.ReadyStateQueue.Remove(tcbCancelar)
		if err != nil {
			logger.Info("No ha sido posible eliminar al TCB con TID: %d de la cola de ReadyStateQueue", tcbCancelar.TID)
			return nil
		}
	}

	if estaEnBlocked {
		err = kernelglobals.BlockedStateQueue.Remove(tcbCancelar)
		if err != nil {
			logger.Info("No ha sido posible eliminar al TCB con TID: %d de la cola de BlockedStateQueue", tcbCancelar.TID)
			return nil
		}
	}

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
		TID: -1, //Para indicar que no hay un hilo en ejecucion puse el -1,
		// total este no afecta en nada y nunca va a haber un hilo con TID -1
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
