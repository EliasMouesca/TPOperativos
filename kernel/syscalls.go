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
	processCreate.PID = types.Pid(PIDcount)
	processCreate.TIDs = []types.Tid{0}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", processCreate.PID)

	kernelsync.ChannelProcessArguments <- args

	return nil
}

func ProcessExit(args []string) error {
	// Esta syscall finalizará el PCB correspondiente al TCB que ejecutó la instrucción,
	// enviando todos sus TCBs asociados a la cola de EXIT. Esta instrucción sólo será llamada por el TID 0
	// del proceso y le deberá indicar a la memoria la finalización de dicho proceso.

	tcb := kernelglobals.ExecStateThread
	pcb := tcb.FatherPCB

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
		// Rami: no deberia ser un if?
		for !kernelglobals.BlockedStateQueue.IsEmpty() {
			blockedTCB, err := kernelglobals.BlockedStateQueue.GetAndRemoveNext()
			if err != nil {
				logger.Error("Error al obtener el siguiente TCB de BlockedStateQueue - %v", err)
				break
			}
			// Si es del PCB que se está finalizando, se mueve a ExitStateQueue
			if blockedTCB.Equal(tcb) {
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

	//kernelsync.ChannelThreadCreate <- args
	//<-kernelsync.SemThreadCreate

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
	logger.Info("Nuevo <%v:%v> agregado a la lista de EveryTCBInTheKernel. ", newTCB.FatherPCB.PID, newTCB.TID)
	currentPCB.TIDs = append(currentPCB.TIDs, newTID)
	logger.Info("El TID: %v fue agregado a la lista de TIDs del PCB: %v. ", newTCB.TID, newTCB.FatherPCB.PID)

	if kernelglobals.ShortTermScheduler == nil {
		logger.Error("ShortTermScheduler no está inicializado.")
		return fmt.Errorf("ShortTermScheduler no inicializado")
	}

	err = kernelglobals.ShortTermScheduler.AddToReady(&kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1])
	logger.Info("## (<%d>:<%d>) Se crea un nuevo hilo - Estado: READY", newTCB.FatherPCB.PID, newTCB.TID)
	if err != nil {
		return fmt.Errorf("error al agregar el TCB a la cola de Ready: %v", err)
	}

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

	//kernelsync.ChannelFinishThread <- 0

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
	for _, mutex := range execPCB.CreatedMutexes {
		if mutex.Name == mutexName {
			encontrado = true
			if mutex.AssignedTCB == nil {
				logger.Info("## El mutex <%v> ha sido asignado al TID <%d> del proceso con PID <%d>",
					mutexName, execTCB.TID, execTCB.FatherPCB.PID)
				mutex.AssignedTCB = execTCB
				execTCB.LockedMutexes = append(execTCB.LockedMutexes, &mutex)
			} else {
				logger.Info("## El mutex <%v> ya está tomado. Bloqueando al TID <%d> del proceso con PID <%d>",
					mutexName, execTCB.TID, execTCB.FatherPCB.PID)
				mutex.BlockedTCBs = append(mutex.BlockedTCBs, execTCB)
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
	// Se deberá verificar primero que exista el mutex solicitado y esté
	// tomado por el hilo que realizó la syscall. En caso de que
	// corresponda, se deberá desbloquear al primer hilo de la cola de
	// bloqueados de ese mutex y le asignará el mutex al hilo recién
	// desbloqueado. Una vez hecho esto, se devuelve la ejecución al hilo
	// que realizó la syscall MUTEX_UNLOCK. En caso de que el hilo que
	// realiza la syscall no tenga asignado el mutex, no realizará
	// ningún desbloqueo.

	mutexName := args[0]
	execTCB := kernelglobals.ExecStateThread
	execPCB := execTCB.FatherPCB

	// TODO: Organizar, for del if del if del if del for del if
	encontrado := false
	for _, mutex := range execPCB.CreatedMutexes {
		if mutex.Name == mutexName {
			encontrado = true
			// Si no está asignado..
			if mutex.AssignedTCB == nil {
				logger.Info("## El hilo actual (TID <%d>) no tiene asignado el mutex <%s>. No se realizará ningún desbloqueo.", execTCB.TID, mutexName)
				return errors.New("el mutex no está asignado a ningún hilo")
			} else {
				if !mutex.AssignedTCB.Equal(execTCB) {
					logger.Debug("Un hilo trató de liberar un mutex que no le fue asignado")
				} else {
					logger.Debug("Liberando mutex <%v> del hilo <%v> del proceso <%v>",
						mutexName, execTCB.TID, execPCB.PID)
					mutex.AssignedTCB = nil
					mutexExistsInTcb := false
					for _, mutexB := range execTCB.LockedMutexes {
						if mutex.Equal(mutexB) {
							mutexExistsInTcb = true

							if len(mutex.BlockedTCBs) > 0 {
								nextTcb := mutex.BlockedTCBs[0]
								mutex.BlockedTCBs = mutex.BlockedTCBs[1:]

								// TODO: Puede darse que nextTcb.LockedMutexes sea null?
								nextTcb.LockedMutexes = append(nextTcb.LockedMutexes, &mutex)
								mutex.AssignedTCB = nextTcb

								err := kernelglobals.ShortTermScheduler.AddToReady(nextTcb)
								if err != nil {
									return err
								}

							} else {
								mutex.AssignedTCB = nil
							}

						}
					}

					if !mutexExistsInTcb {
						logger.Error("El mutex apuntaba a un TCB que lo lockeó y el TCB no sabía que tenía asignado ese mutex!!! Esto no debería pasar nunca")
					}

				} // End if mutex tiene asignado este hilo

			} // End else el mutex tiene un hilo que lo lockear
		}
	}

	if !encontrado {
		return errors.New(fmt.Sprintf("No se encontró el mutex <%v>", mutexName))
	}

	return nil
}
