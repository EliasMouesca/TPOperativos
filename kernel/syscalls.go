package main

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
	"time"
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
	syscalls.DumpMemory:    DumpMemory,
	syscalls.IO:            IO,
}

var PIDcount int = 0

func ProcessCreate(args []string) error {
	// Esta syscall recibirá 3 parámetros de la CPU: nombre del archivo, tamaño del proceso y prioridad del hilo main (TID 0).
	// El Kernel creará un nuevo PCB y lo dejará en estado NEW.

	// Se crea el PCB (sin crear el hilo principal aún)
	var processCreate kerneltypes.PCB
	PIDcount++
	processCreate.PID = types.Pid(PIDcount)
	processCreate.TIDs = []types.Tid{0} // Solo se conoce el TID 0 por ahora

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", processCreate.PID)

	// Agregar el PCB a la lista de PCBs en el kernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, processCreate)

	// Mandar el proceso a la cola de NewStateQueue (solo PCB, sin TCB)
	kernelglobals.NewPCBStateQueue.Add(&kernelglobals.EveryPCBInTheKernel[len(kernelglobals.EveryPCBInTheKernel)-1])

	//Agrego el PID a args, para despues pasarselo a memoria
	pidStr := strconv.Itoa(int(processCreate.PID))
	args = append(args, pidStr)

	// Enviar los argumentos al canal para que NewProcessToReady los procese
	kernelsync.ChannelProcessArguments <- args

	return nil
}

func ProcessExit(args []string) error {
	// Esta syscall finalizará el PCB correspondiente al TCB que ejecutó la instrucción,
	// enviando todos sus TCBs asociados a la cola de EXIT. Esta instrucción sólo será llamada por el TID 0
	// del proceso y le deberá indicar a la memoria la finalización de dicho proceso.

	tcb := kernelglobals.ExecStateThread
	pcb := tcb.FatherPCB

	// Verificar que el hilo que llama sea el main (TID 0)
	if tcb.TID != 0 {
		return errors.New("el hilo que quiso eliminar el proceso no es el hilo main")
	}

	// Eliminar todos los hilos del PCB de las colas de Ready
	for _, tid := range pcb.TIDs {
		// 1. Verificar y eliminar hilos de la cola de Ready
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
			if blockedTCB.FatherPCB.PID == pcb.PID {
				kernelglobals.ExitStateQueue.Add(blockedTCB)
				logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de BlockedStateQueue y se movió a ExitStateQueue", blockedTCB.TID, pcb.PID)
			} else {
				// Si no es, se vuelve a insertar en la cola de bloqueados
				kernelglobals.BlockedStateQueue.Add(blockedTCB)
			}
		}

		// 3. Verificar y eliminar hilos en la cola de New
		for !kernelglobals.NewStateQueue.IsEmpty() {
			newTCB, err := kernelglobals.NewStateQueue.GetAndRemoveNext()
			if err != nil {
				logger.Error("Error al obtener el siguiente TCB de NewStateQueue - %v", err)
				break
			}
			// Si es del PCB que se está finalizando, se mueve a ExitStateQueue
			if newTCB.FatherPCB.PID == pcb.PID {
				kernelglobals.ExitStateQueue.Add(newTCB)
				logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de NewStateQueue y se movió a ExitStateQueue", newTCB.TID, pcb.PID)
			} else {
				// Si no es, se vuelve a insertar en la cola de new
				kernelglobals.NewStateQueue.Add(newTCB)
			}
		}
	}

	// Finalmente, mover el hilo principal (ExecStateThread) a ExitStateQueue
	kernelglobals.ExitStateQueue.Add(tcb)
	kernelglobals.ExecStateThread = nil

	// Enviar la señal a la memoria sobre la finalización del proceso
	kernelsync.ChannelFinishprocess <- pcb.PID

	logger.Info("## Finaliza el proceso <%v>", pcb.PID)

	return nil
}

func ThreadCreate(args []string) error {

	// Esta syscall recibirá como parámetro de la CPU el nombre del archivo de
	// pseudocódigo que deberá ejecutar el hilo a crear y su prioridad. Al momento de crear el nuevo hilo,
	// deberá generar el nuevo TCB con un TID autoincremental y poner al mismo en el estado READY.

	prioridad, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("error al convertir la prioridad a entero: %v", err)
	}

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.FatherPCB

	newTID := types.Tid(len(currentPCB.TIDs))

	newTCB := kerneltypes.TCB{
		TID:       newTID,
		Prioridad: prioridad,
		FatherPCB: currentPCB,
	}

	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, newTCB)
	logger.Info("Nuevo (<%v:%v>) agregado a la lista de EveryTCBInTheKernel. ", newTCB.FatherPCB.PID, newTCB.TID)
	currentPCB.TIDs = append(currentPCB.TIDs, newTID)
	logger.Info("El TID: %v fue agregado a la lista de TIDs del PCB: %v. ", newTCB.TID, newTCB.FatherPCB.PID)

	/*
		if kernelglobals.ShortTermScheduler == nil {
			logger.Error("ShortTermScheduler no está inicializado.")
			return fmt.Errorf("ShortTermScheduler no inicializado")
		}

		err = kernelglobals.ShortTermScheduler.AddToReady(&kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1])
		logger.Info("## (<%d>:<%d>) Se crea un nuevo hilo - Estado: READY", newTCB.FatherPCB.PID, newTCB.TID)
		if err != nil {
			return fmt.Errorf("error al agregar el TCB a la cola de Ready: %v", err)
		}*/

	kernelglobals.NewStateQueue.Add(&kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1])
	logger.Info("(<%v:%v>) fue agregado a NewStateQueue.", newTCB.FatherPCB.PID, newTCB.TID)

	kernelsync.ChannelThreadCreate <- args

	return nil
}

func ThreadJoin(args []string) error {
	// Esta syscall recibe como parámetro un TID, mueve el hilo que la invocó al estado
	// BLOCK hasta que el TID pasado por parámetro finalice. En caso de que el TID pasado por parámetro
	// no exista o ya haya finalizado, esta syscall no hace nada y el hilo que la invocó continuará su
	// ejecución.

	tidString, err := strconv.Atoi(args[0])
	tidToJoin := types.Tid(tidString)

	if err != nil {
		return errors.New("error al convertir el TID a entero")
	}
	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.FatherPCB

	finalizado := false
	queueSize := kernelglobals.ExitStateQueue.Size()
	for i := 0; i < queueSize; i++ {
		tcb, err := kernelglobals.ExitStateQueue.GetAndRemoveNext()
		if err != nil {
			return errors.New("error al obtener el siguiente TCB de ExitStateQueue")
		}
		if tcb.TID == tidToJoin && tcb.FatherPCB == currentPCB {
			finalizado = true
		}
		kernelglobals.ExitStateQueue.Add(tcb)
	}

	if finalizado {
		logger.Info("## TID <%d> ya ha finalizado. Continúa la ejecución de (<%v>:<%v>).", currentPCB.PID, execTCB.TID, tidToJoin)
		return nil
	}

	tidExiste := false
	for _, tid := range currentPCB.TIDs {
		if tid == tidToJoin {
			tidExiste = true
			break
		}
	}

	if !tidExiste {
		logger.Info("## (<%d>:<%d>) TID <%d> no pertenece a la lista de TIDs del PCB con PID <%d>. Continúa la ejecución.",
			currentPCB.PID,
			execTCB.TID,
			tidToJoin,
			currentPCB.PID)
		return nil
	}

	for _, tcbToJoin := range kernelglobals.EveryTCBInTheKernel {
		if tcbToJoin.TID == tidToJoin && tcbToJoin.FatherPCB.Equal(execTCB.FatherPCB) {
			execTCB.JoinedTCB = &tcbToJoin
			logger.Info("Estado del atributo JoinedTCB del TCB que llamo a ThreadJoin luego de ejecutar: %v", execTCB.JoinedTCB)
		}
	}

	kernelglobals.BlockedStateQueue.Add(execTCB)

	kernelglobals.ExecStateThread = nil

	logger.Info("## (<%d>:<%d>) Hilo se mueve a estado BLOCK esperando a TID <%d>", currentPCB.PID, execTCB.TID, tidToJoin)
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

	currentPCB := kernelglobals.ExecStateThread.FatherPCB

	// Intentar eliminar el TID de las colas Ready usando ThreadRemove del planificador
	err = kernelglobals.ShortTermScheduler.ThreadRemove(types.Tid(tidCancelar), currentPCB.PID)
	if err == nil {
		logger.Info("Se movió el TID <%d> del PCB con PID <%d> de ReadyStateQueue a ExitStateQueue", tidCancelar, currentPCB.PID)
		return nil
	}

	// Si no estaba en Ready, verificar y eliminar hilos en la cola de Blocked
	for !kernelglobals.BlockedStateQueue.IsEmpty() {
		tcb, err := kernelglobals.BlockedStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error al obtener el siguiente TCB de BlockedStateQueue: %v", err)
			break
		}

		if tcb.TID == types.Tid(tidCancelar) && tcb.FatherPCB == currentPCB {
			kernelglobals.ExitStateQueue.Add(tcb)
			logger.Info("Se movió el TID <%d> del PCB con PID <%d> de BlockedStateQueue a ExitStateQueue", tidCancelar, currentPCB.PID)
			return nil
		} else {
			kernelglobals.BlockedStateQueue.Add(tcb)
		}
	}

	logger.Info("## No se encontró el TID <%d> en ninguna cola para el PCB con PID <%d>. Continúa la ejecución normal.", tidCancelar, currentPCB.PID)
	return nil
}

func ThreadExit(args []string) error {
	// Esta syscall finaliza al hilo que la invocó, pasando el mismo al estado EXIT.
	// Se deberá indicar a la Memoria la finalización de dicho hilo.

	execTCB := kernelglobals.ExecStateThread

	kernelsync.ChannelFinishThread <- []string{strconv.Itoa(int(execTCB.TID)), strconv.Itoa(int(execTCB.FatherPCB.PID))}

	kernelsync.SemFinishThread <- struct{}{}
	<-kernelsync.SemMovedFinishThreads

	logger.Info("## Moviendo el TID <%d> al estado EXIT", execTCB.TID)
	kernelglobals.ExitStateQueue.Add(execTCB)

	kernelglobals.ExecStateThread = nil

	return nil
}

func MutexCreate(args []string) error {
	// TODO: chequear numero de args
	// TODO: Chequear que el nombre sea único
	// Crea un nuevo mutex para el proceso sin asignar a ningún hilo.

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.FatherPCB

	newMutex := kerneltypes.Mutex{
		Name:        args[0],
		AssignedTCB: nil,
		BlockedTCBs: []*kerneltypes.TCB{},
	}

	currentPCB.CreatedMutexes = append(currentPCB.CreatedMutexes, newMutex)

	logger.Info("## Se creó el mutex <%v> para el proceso con PID <%d>", newMutex.Name, currentPCB.PID)

	return nil
}

func MutexLock(args []string) error {

	mutexName := args[0]

	execTCB := kernelglobals.ExecStateThread
	execPCB := execTCB.FatherPCB

	encontrado := false
	for i := range execPCB.CreatedMutexes {
		mutex := &execPCB.CreatedMutexes[i]
		if mutex.Name == mutexName {
			encontrado = true
			if mutex.AssignedTCB == nil {
				logger.Info("## El mutex <%v> ha sido asignado al TID <%d> del proceso con PID <%d>", mutexName, execTCB.TID, execTCB.FatherPCB.PID)
				mutex.AssignedTCB = execTCB
				execTCB.LockedMutexes = append(execTCB.LockedMutexes, mutex)
				logger.Info("AssignedTCB del Mutex %s: %v", mutexName, mutex.AssignedTCB)
				logger.Info("LockedMutexes del TCB %v: %v", execTCB, execTCB.LockedMutexes)
			} else {
				logger.Info("## El mutex <%v> ya está tomado. Bloqueando al TID <%d> del proceso con PID <%d>", mutexName, execTCB.TID, execTCB.FatherPCB.PID)
				mutex.BlockedTCBs = append(mutex.BlockedTCBs, execTCB)
				kernelglobals.BlockedStateQueue.Add(execTCB)
			}
		}
	}
	if !encontrado {
		logger.Debug("Se pidió un mutex no existía")
		return errors.New(fmt.Sprintf("No se encontró el mutex <%v>", mutexName))
	}

	return nil
}

func MutexUnlock(args []string) error {
	mutexName := args[0]
	execTCB := kernelglobals.ExecStateThread
	execPCB := execTCB.FatherPCB

	encontrado := false
	for i := range execPCB.CreatedMutexes {
		mutex := &execPCB.CreatedMutexes[i]

		if mutex.Name == mutexName {
			logger.Info("Se ha encontrado el mutex que se desea realizar UnLock.")
			encontrado = true

			if mutex.AssignedTCB == nil {
				logger.Info("## El hilo actual (TID <%d>) no tiene asignado el mutex <%s>. No se realizará ningún desbloqueo.", execTCB.TID, mutexName)
				return errors.New("el mutex no está asignado a ningún hilo")
			}

			if mutex.AssignedTCB.TID != execTCB.TID {
				logger.Debug("Un hilo trató de liberar un mutex que no le fue asignado")
				return nil
			}

			logger.Info("Liberando mutex <%v> del hilo <%v> del proceso <%v>", mutexName, execTCB.TID, execPCB.PID)
			mutex.AssignedTCB = nil

			// Remover el mutex de la lista LockedMutexes del hilo actual de manera segura
			for i, lockedMutex := range execTCB.LockedMutexes {
				if lockedMutex.Equal(mutex) {
					logger.Info("Removiendo mutex <%s> de la lista LockedMutexes del TCB <%d>", mutexName, execTCB.TID)
					// Eliminar el mutex correctamente de la lista
					execTCB.LockedMutexes = append(execTCB.LockedMutexes[:i], execTCB.LockedMutexes[i+1:]...)
					break
				}
			}

			// Si hay hilos bloqueados en este mutex, desbloquear el primero
			if len(mutex.BlockedTCBs) > 0 {
				nextTcb := mutex.BlockedTCBs[0]
				mutex.BlockedTCBs = mutex.BlockedTCBs[1:]

				// Asegurarse de que la lista LockedMutexes esté inicializada
				if nextTcb.LockedMutexes == nil {
					nextTcb.LockedMutexes = []*kerneltypes.Mutex{}
				}

				nextTcb.LockedMutexes = append(nextTcb.LockedMutexes, mutex)
				mutex.AssignedTCB = nextTcb

				err := kernelglobals.ShortTermScheduler.AddToReady(nextTcb)
				if err != nil {
					return err
				}
			}

			return nil
		}
	}

	if !encontrado {
		return errors.New(fmt.Sprintf("No se encontró el mutex <%v>", mutexName))
	}

	return nil
}

func DumpMemory(args []string) error {
	// Obtener el thread ejecutándose
	execTCB := kernelglobals.ExecStateThread
	if execTCB == nil {
		return fmt.Errorf("no hay un hilo en ejecución")
	}
	pcb := execTCB.FatherPCB

	pid := strconv.Itoa(int(execTCB.FatherPCB.PID))
	tid := strconv.Itoa(int(execTCB.TID))

	// Mover el hilo actual a la cola de bloqueados antes de hacer el request a memoria
	logger.Info("Moviendo el hilo a la cola de bloqueados (PID: %v, TID: %v)", pid, tid)
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue.Add(execTCB)

	// Crear el request para la memoria
	request := types.RequestToMemory{
		Type:      types.MemoryDump,
		Arguments: []string{pid, tid},
	}

	// Enviar request a memoria
	err := sendMemoryRequest(request)
	if err != nil {
		logger.Error("Error en el request a memoria para DumpMemory - %v", err)

		// Mover el proceso a estado EXIT en caso de error
		kernelglobals.BlockedStateQueue.Remove(execTCB) // Quitar de la cola de bloqueados
		kernelglobals.ExitStateQueue.Add(execTCB)

		// Eliminar todos los hilos del PCB de las colas de Ready
		for _, tid := range pcb.TIDs {
			// 1. Verificar y eliminar hilos de la cola de Ready
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
				if blockedTCB.FatherPCB.PID == pcb.PID {
					kernelglobals.ExitStateQueue.Add(blockedTCB)
					logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de BlockedStateQueue y se movió a ExitStateQueue", blockedTCB.TID, pcb.PID)
				} else {
					// Si no es, se vuelve a insertar en la cola de bloqueados
					kernelglobals.BlockedStateQueue.Add(blockedTCB)
				}
			}

			// 3. Verificar y eliminar hilos en la cola de New
			for !kernelglobals.NewStateQueue.IsEmpty() {
				newTCB, err := kernelglobals.NewStateQueue.GetAndRemoveNext()
				if err != nil {
					logger.Error("Error al obtener el siguiente TCB de NewStateQueue - %v", err)
					break
				}
				// Si es del PCB que se está finalizando, se mueve a ExitStateQueue
				if newTCB.FatherPCB.PID == pcb.PID {
					kernelglobals.ExitStateQueue.Add(newTCB)
					logger.Info("Se eliminó el TID <%d> del PCB con PID <%d> de NewStateQueue y se movió a ExitStateQueue", newTCB.TID, pcb.PID)
				} else {
					// Si no es, se vuelve a insertar en la cola de new
					kernelglobals.NewStateQueue.Add(newTCB)
				}
			}
		}

		// Limpiar el hilo en ejecución
		kernelglobals.ExecStateThread = nil
		logger.Info("El proceso con PID <%v> y TID <%v> fue movido a EXIT por error en DumpMemory", pid, tid)

		return err
	}

	// Si la operación fue exitosa, mover el hilo de bloqueados a la cola de READY
	kernelglobals.BlockedStateQueue.Remove(execTCB)      // Quitar de la cola de bloqueados
	kernelglobals.ShortTermScheduler.AddToReady(execTCB) // Mover a READY
	logger.Info("DumpMemory completado exitosamente, moviendo hilo a READY (PID: %v, TID: %v)", pid, tid)

	return nil
}

func IO(args []string) error {
	kernelsync.MutexPlanificadorLP.Lock()
	threadBlockedTime, _ := strconv.Atoi(args[0])
	execTCB := kernelglobals.ExecStateThread

	for kernelglobals.ExecStateThread == nil {
	}
	kernelglobals.BlockedStateQueue.Add(execTCB)

	// Canal FIFO
	logger.Info("Bloqueando el hilo %v de duracion %v", execTCB.TID, threadBlockedTime)
	kernelsync.MutexPlanificadorLP.Unlock()

	kernelsync.ChannelIO <- execTCB

	time.Sleep(time.Duration(threadBlockedTime) * time.Millisecond)

	tcbBlock := <-kernelsync.ChannelIO
	err := kernelglobals.BlockedStateQueue.Remove(tcbBlock)
	if err != nil {
		logger.Error("No se pudo remover el tcb de la BlockQueue - %v", err)
	}
	logger.Info("Desbloqueando el hilo %v", tcbBlock)

	return nil
}
