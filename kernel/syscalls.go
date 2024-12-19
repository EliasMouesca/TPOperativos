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
	go func() error {
		// Si hay un proceso en espera para ser creado, tiene que esperar a que ese se cree
		kernelsync.MutexPuedoCrearProceso.Lock()
		// Se crea el PCB (sin crear el hilo principal aún)
		var processCreate kerneltypes.PCB
		PIDcount++
		processCreate.PID = types.Pid(PIDcount)
		processCreate.TIDs = []types.Tid{0} // Solo se conoce el TID 0 por ahora
		processCreate.CreatedMutexes = []kerneltypes.Mutex{}

		logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", processCreate.PID)

		// Agregar el PCB a la lista de PCBs en el kernel
		kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, processCreate)

		// Buscar el PCB recien creado utilizando el PID
		pcbPtr := buscarPCBPorPID(processCreate.PID)
		if pcbPtr == nil {
			logger.Error("No se encontró el PCB con PID <%d> en la lista global", processCreate.PID)
			return errors.New("PCB no encontrado")
		}

		// Mandar el proceso a la cola de NewStateQueue (solo PCB, sin TCB)
		kernelglobals.NewPCBStateQueue.Add(pcbPtr)

		//Agrego el PID a args, para despues pasarselo a memoria
		pidStr := strconv.Itoa(int(processCreate.PID))
		args = append(args, pidStr)

		// Enviar los argumentos al canal para que NewProcessToReady los procese
		kernelsync.ChannelProcessArguments <- args
		go func() {
			<-kernelsync.SemProcessCreateOK
		}()

		return nil
	}()

	return nil
}

func ProcessExit(args []string) error {
	// Esta syscall finalizará el PCB correspondiente al TCB que ejecutó la instrucción,
	// enviando todos sus TCBs asociados a la cola de EXIT. Esta instrucción sólo será llamada por el TID 0
	// del proceso y le deberá indicar a la memoria la finalización de dicho proceso.
	tcb := kernelglobals.ExecStateThread
	pcb := tcb.FatherPCB

	if tcb.TID != 0 {
		// Verificar que el hilo que llama sea el main (TID 0)
		return errors.New("el hilo que quiso eliminar el proceso no es el hilo main")
	}

	// Enviar la señal a la memoria sobre la finalización del proceso
	logger.Debug("Entra a process exit")

	kernelsync.ChannelFinishprocess <- pcb.PID
	<-kernelsync.ChannelFinishProcess2
	// Eliminar todos los hilos del PCB de las colas de Ready
	for _, tid := range pcb.TIDs {
		// 1. Verificar y eliminar hilos de la cola de Ready
		existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(tid, pcb.PID)
		if existsInReady {
			err := kernelglobals.ShortTermScheduler.ThreadRemove(tid, pcb.PID)
			//kernelglobals.ExitStateQueue.Add(tcb)
			err = agregarAExitStateQueue(tcb)
			if err != nil {
				logger.Error("Error al eliminar el TID <%d> del PCB con PID <%d> de las colas de Ready - %v", tid, pcb.PID, err)
			}
			<-kernelsync.PendingThreadsChannel
			logger.Info("## (<%d:%d>) Se quita de ready", pcb.PID, tid)
		}
	}

	if kernelglobals.BlockedStateQueue.Contains(tcb) {
		kernelglobals.BlockedStateQueue.Remove(tcb)
		logger.Info("(<%d:%d>) Se quita de blocked el hilo", pcb.PID, tcb.TID)
	}

	if kernelglobals.NewStateQueue.Contains(tcb) {
		kernelglobals.NewStateQueue.Remove(tcb)
		logger.Info("(<%d:%d>) Se manda a exit", pcb.PID, tcb.TID)
	}

	// Finalmente, mover el hilo principal (ExecStateThread) a ExitStateQueue
	kernelglobals.ExitStateQueue.Add(tcb)
	kernelglobals.ExecStateThread = nil

	logger.Info("## (<%v>) Finaliza el proceso", pcb.PID)

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

	currentPCB.TIDs = append(currentPCB.TIDs, newTID)

	for i := range kernelglobals.EveryPCBInTheKernel {
		if kernelglobals.EveryPCBInTheKernel[i].PID == currentPCB.PID {
			kernelglobals.EveryPCBInTheKernel[i] = *currentPCB
			break
		}
	}

	kernelglobals.NewStateQueue.Add(buscarTCBPorTID(newTID, currentPCB.PID))

	kernelsync.ChannelThreadCreate <- args
	<-kernelsync.ThreadCreateComplete

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
		//kernelglobals.ExitStateQueue.Add(tcb)
		agregarAExitStateQueue(tcb)
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

	// Buscar el TCB del TID a joinear
	tcbToJoin := buscarTCBPorTID(tidToJoin, currentPCB.PID)
	if tcbToJoin == nil {
		return errors.New("no se encontró el TCB del hilo a joinear en EveryTCBInTheKernel")
	}

	// Modificar execTCB para que tenga el puntero a tcbToJoin
	execTCB.JoinedTCB = tcbToJoin

	for i := range kernelglobals.EveryTCBInTheKernel {
		if kernelglobals.EveryTCBInTheKernel[i].TID == execTCB.TID && kernelglobals.EveryTCBInTheKernel[i].FatherPCB.PID == execTCB.FatherPCB.PID {
			kernelglobals.EveryTCBInTheKernel[i] = *execTCB
			break
		}
	}

	kernelglobals.BlockedStateQueue.Add(execTCB)
	logger.Info("## (<%v>:<%v>)- Bloqueado por: <PTHREAD_JOIN>", currentPCB.PID, execTCB.TID)
	logger.Debug("(%v : %v)- PTHREAD_JOIN a (%v : %v) ", currentPCB.PID, execTCB.TID, tcbToJoin.FatherPCB.PID, tcbToJoin.TID)
	kernelglobals.ExecStateThread.QuantumRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
	kernelglobals.ExecStateThread = nil
	logger.Info("## (<%v>:<%v>) Se saco de Exec", currentPCB.PID, execTCB.TID)

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
	//kernelglobals.ExitStateQueue.Add(buscarTCBPorTID(types.Tid(tidCancelar), currentPCB.PID))
	agregarAExitStateQueue(buscarTCBPorTID(types.Tid(tidCancelar), currentPCB.PID))
	if err == nil {
		logger.Info("## (<%d:%d>) Finaliza el hilo", currentPCB.PID, tidCancelar)
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
			//kernelglobals.ExitStateQueue.Add(tcb)
			agregarAExitStateQueue(tcb)
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

	<-kernelsync.ThreadExitComplete

	//kernelglobals.ExitStateQueue.Add(execTCB)
	agregarAExitStateQueue(execTCB)
	kernelglobals.ExecStateThread = nil
	logger.Info("## (<%v>:<%v>) Finaliza el hilo", execTCB.FatherPCB.PID, execTCB.TID)

	return nil
}

func MutexCreate(args []string) error {
	// Crea un nuevo mutex para el proceso sin asignar a ningún hilo.

	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.FatherPCB

	for _, existingMutex := range currentPCB.CreatedMutexes {
		if existingMutex.Name == args[0] {
			return fmt.Errorf("ya existe un mutex con el nombre <%v> en el proceso con PID <%d>", args[0], currentPCB.PID)
		}
	}

	newMutex := kerneltypes.Mutex{
		Name:        args[0],
		AssignedTCB: nil,
		BlockedTCBs: []*kerneltypes.TCB{},
	}

	currentPCB.CreatedMutexes = append(currentPCB.CreatedMutexes, newMutex)

	// Ahora buscar el PCB en la lista de EveryPCBInTheKernel y actualizarlo
	//for i, pcb := range kernelglobals.EveryPCBInTheKernel {
	//	if pcb.PID == currentPCB.PID {
	//		kernelglobals.EveryPCBInTheKernel[i] = *currentPCB
	//		break
	//	}
	//}

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
				mutex.AssignedTCB = execTCB
				execTCB.LockedMutexes = append(execTCB.LockedMutexes, mutex)
				logger.Info("## El mutex <%v> ha sido asignado a (<%d:%d>)", mutexName, execTCB.FatherPCB.PID, execTCB.TID)

			} else {
				logger.Info("“## (<%v>:<%v>)- Bloqueado por: <MUTEX> (nombre: %v)", execTCB.FatherPCB.PID, execTCB.TID, mutexName)
				mutex.BlockedTCBs = append(mutex.BlockedTCBs, execTCB)
				kernelglobals.ShortTermScheduler.ThreadRemove(execTCB.TID, execTCB.FatherPCB.PID)
				kernelglobals.BlockedStateQueue.Add(execTCB)

				kernelglobals.ExecStateThread.QuantumRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
				kernelglobals.ExecStateThread = nil
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

	//pid := strconv.Itoa(int(execTCB.FatherPCB.PID))
	//tid := strconv.Itoa(int(execTCB.TID))
	pid := execTCB.FatherPCB.PID
	tid := execTCB.TID
	// Mover el hilo actual a la cola de bloqueados antes de hacer el request a memoria
	logger.Info("## (<%v:%v>) Se movio a la cola de Bloqueados", pid, tid)
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue.Add(execTCB)

	// Crear el request para la memoria
	request := types.RequestToMemory{
		Type:   types.MemoryDump,
		Thread: types.Thread{PID: pid, TID: tid},
	}

	// Enviar request a memoria
	err := sendMemoryRequest(request)
	if err != nil {
		logger.Error("Error en el request a memoria para DumpMemory - %v", err)

		// Mover el proceso a estado EXIT en caso de error
		kernelglobals.BlockedStateQueue.Remove(execTCB) // Quitar de la cola de bloqueados
		//kernelglobals.ExitStateQueue.Add(execTCB)
		agregarAExitStateQueue(execTCB)

		// Eliminar todos los hilos del PCB de las colas de Ready
		for _, tid := range pcb.TIDs {
			// 1. Verificar y eliminar hilos de la cola de Ready
			existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(tid, pcb.PID)
			if existsInReady {
				err := kernelglobals.ShortTermScheduler.ThreadRemove(tid, pcb.PID)
				//kernelglobals.ExitStateQueue.Add(buscarTCBPorTID(tid, pcb.PID))
				agregarAExitStateQueue(buscarTCBPorTID(tid, pcb.PID))
				if err != nil {
					logger.Error("Error al eliminar el TID <%d> del PCB con PID <%d> de las colas de Ready - %v", tid, pcb.PID, err)
				} else {
					logger.Info("## (<%v:%v>) Finaliza el hilo", pcb.PID, tid)
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
					//kernelglobals.ExitStateQueue.Add(blockedTCB)
					agregarAExitStateQueue(blockedTCB)
					logger.Info("## (<%v:%v>) Finaliza el hilo", pcb.PID, blockedTCB.TID)
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
					//kernelglobals.ExitStateQueue.Add(newTCB)
					agregarAExitStateQueue(newTCB)
					logger.Info("## (<%v:%v>) Finaliza el hilo", pcb.PID, newTCB.TID)
				} else {
					// Si no es, se vuelve a insertar en la cola de new
					kernelglobals.NewStateQueue.Add(newTCB)
				}
			}
		}

		// Limpiar el hilo en ejecución
		kernelglobals.ExecStateThread.QuantumRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
		kernelglobals.ExecStateThread = nil
		logger.Info("El proceso con PID <%v> y TID <%v> fue movido a EXIT por error en DumpMemory", pid, tid)

		return err
	}

	// Si la operación fue exitosa, mover el hilo de bloqueados a la cola de READY
	kernelglobals.BlockedStateQueue.Remove(execTCB)      // Quitar de la cola de bloqueados
	kernelglobals.ShortTermScheduler.AddToReady(execTCB) // Mover a READY
	logger.Info("DumpMemory completado exitosamente! Moviendo (<%v:%v>) a READY", pid, tid)

	return nil
}

func IO(args []string) error {
	threadBlockedTime, _ := strconv.Atoi(args[0])
	execTCB := kernelglobals.ExecStateThread

	kernelglobals.BlockedStateQueue.Add(execTCB)
	// Canal FIFO
	logger.Info("## (<%v>:<%v>) - Bloqueado por: <IO>", execTCB.FatherPCB.PID, execTCB.TID)

	kernelglobals.ExecStateThread.QuantumRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
	kernelglobals.ExecStateThread = nil
	go func() {
		logger.Debug("Antes de pasar argumentos al ChannelIO")
		kernelsync.ChannelIO <- execTCB
		kernelsync.ChannelIO2 <- threadBlockedTime
		logger.Debug("Despues de pasar argumentos al ChannelIO")
	}()

	return nil
}

func buscarPCBPorPID(pid types.Pid) *kerneltypes.PCB {
	for i := range kernelglobals.EveryPCBInTheKernel {
		if kernelglobals.EveryPCBInTheKernel[i].PID == pid {
			return &kernelglobals.EveryPCBInTheKernel[i]
		}
	}
	return nil
}
func buscarTCBPorTID(tid types.Tid, pid types.Pid) *kerneltypes.TCB {
	for i := range kernelglobals.EveryTCBInTheKernel {
		if kernelglobals.EveryTCBInTheKernel[i].TID == tid && kernelglobals.EveryTCBInTheKernel[i].FatherPCB.PID == pid {
			return &kernelglobals.EveryTCBInTheKernel[i]
		}
	}
	return nil
}

func agregarAExitStateQueue(tcb *kerneltypes.TCB) error {
	logger.Debug("Buscando: (%v : %v)", tcb.FatherPCB.PID, tcb.TID)
	logger.Debug("EvereyTCBInTheKernel")

	for _, tcb1 := range kernelglobals.EveryTCBInTheKernel {
		logger.Debug("(%v: %v)", tcb1.FatherPCB.PID, tcb1.TID)
	}
	estaEnExit := false
	for _, tcb1 := range kernelglobals.ExitStateQueue.GetElements() {
		if tcb1.Equal(tcb) {
			estaEnExit = true
			break
		}
	}
	if !estaEnExit {
		kernelglobals.ExitStateQueue.Add(tcb)
		for i, tid := range tcb.FatherPCB.TIDs {
			if tid == tcb.TID {
				tcb.FatherPCB.TIDs = append(tcb.FatherPCB.TIDs[:i], tcb.FatherPCB.TIDs[i+1:]...)
				break
			}
		}
	} else {
		logger.Warn("ya estaba en exit y se quizo volver a agregar")
	}
	return nil
}
