package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func planificadorLargoPlazo() {
	// En el enunciado en implementacion dice que hay que inicializar un proceso
	// quiza hay que hacerlo aca o en kernel.go es lo mismo creo

	logger.Info("Iniciando el planificador de largo plazo")

	go NewProcessToReady()
	go ProcessToExit()
	go NewThreadToReady()
	go ThreadToExit()

}

func NewProcessToReady() {
	for {
		// Espera los argumentos del proceso desde el canal
		args := <-kernelsync.ChannelProcessArguments
		logger.Debug("Llegaron los argumentos de la syscall: %v", args)
		fileName := args[0]
		processSize := args[1]
		prioridad, _ := strconv.Atoi(args[2])
		pid, _ := strconv.Atoi(args[3])

		// Crear el request para verificar si memoria tiene espacio
		request := types.RequestToMemory{
			Thread:    types.Thread{PID: types.Pid(pid)},
			Type:      types.CreateProcess,
			Arguments: []string{fileName, processSize},
		}

		logger.Debug("Preguntando a memoria si tiene espacio disponible")
		// Loop hasta que memoria confirme que tiene espacio
		for {
			err := sendMemoryRequest(request)
			if err != nil {
				logger.Error("Error al enviar request a memoria: %v", err)
				<-kernelsync.InitProcess // Espera a que finalice otro proceso antes de intentar de nuevo
			} else {
				logger.Debug("Hay espacio disponible en memoria")
				break
			}
		}

		// Obtener el PCB desde la cola de NewStateQueue
		pcb, err := kernelglobals.NewPCBStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error en la cola NEW - %v", err)
			continue
		}

		// Crear el hilo principal (mainThread) ahora que el proceso tiene espacio en memoria
		mainThread := kerneltypes.TCB{
			TID:       0,
			Prioridad: prioridad,
			FatherPCB: pcb,
		}

		// Agregar el mainThread a la lista de TCBs en el kernel
		kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread)

		// Obtener el puntero del hilo principal para encolarlo en Ready
		mainThreadPtr := &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]

		// Mover el mainThread a la cola de Ready
		kernelglobals.ShortTermScheduler.AddToReady(mainThreadPtr)
		logger.Info("Se agregó el hilo main (TID 0) del proceso PID <%d> a la cola Ready", pcb.PID)
	}
}

func ProcessToExit() {
	for {
		// Recibir la señal de finalización de un proceso
		PID := <-kernelsync.ChannelFinishprocess

		// Simular la comunicación con memoria (puede ser mockeado o real)
		request := types.RequestToMemory{
			Thread:    types.Thread{PID: PID},
			Type:      types.FinishProcess,
			Arguments: []string{},
		}
		logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", PID)

		// Simular el request a memoria (puede reemplazarse con la función real si está disponible)
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error en el request a memoria: %v", err)
		}
		kernelsync.InitProcess <- 0
	}
}

func NewThreadToReady() {
	for {
		// Recibir los argumentos a través del canal
		args := <-kernelsync.ChannelThreadCreate
		fileName := args[0]

		// Tomar el siguiente TCB de la cola NewStateQueue
		newTCB, err := kernelglobals.NewStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error al obtener el siguiente TCB de NewStateQueue: %v", err)
			continue
		}

		// Informar a memoria sobre la creación del hilo
		request := types.RequestToMemory{
			Thread:    types.Thread{PID: newTCB.FatherPCB.PID, TID: newTCB.TID},
			Type:      types.CreateThread,
			Arguments: []string{fileName},
		}
		logger.Debug("Informando a Memoria sobre la creación de un hilo")

		// Enviar la solicitud a memoria
		err = sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error en el request a memoria: %v", err)
			continue
		}

		// Una vez confirmada la creación, agregar el TCB a la cola de Ready
		err = kernelglobals.ShortTermScheduler.AddToReady(newTCB)
		if err != nil {
			logger.Error("Error al agregar el TCB a la cola de Ready: %v", err)
			continue
		}
		logger.Info("## (<%d>:<%d>) Se movió el hilo de NEW a READY", newTCB.FatherPCB.PID, newTCB.TID)

		// Notificar que se completó la creación del hilo
		// kernelsync.SemThreadCreate <- 0
	}
}

func ThreadToExit() {
	// Al momento de finalizar un hilo, el Kernel deberá informar a la Memoria
	// la finalización del mismo y deberá mover al estado READY a todos los
	// hilos que se encontraban bloqueados por ese TID. De esta manera, se
	// desbloquean aquellos hilos bloqueados por THREAD_JOIN y por mutex
	// tomados por el hilo finalizado (en caso que hubiera).

	for {
		// Leer los argumentos enviados por ThreadExit
		args := <-kernelsync.ChannelFinishThread
		tid, err := strconv.Atoi(args[0])
		if err != nil {
			logger.Error("Error al convertir el TID: %v", err)
			continue
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			logger.Error("Error al convertir el PID: %v", err)
			continue
		}

		<-kernelsync.SemFinishThread

		logger.Info("## Iniciando finalización del TID <%v> del PCB con PID <%d>", tid, pid)

		// Obtener el TCB correspondiente del kernel
		var execTCB *kerneltypes.TCB
		for _, tcb := range kernelglobals.EveryTCBInTheKernel {
			if int(tcb.TID) == tid && int(tcb.FatherPCB.PID) == pid {
				execTCB = &tcb
				break
			}
		}
		if execTCB == nil {
			logger.Error("No se encontró el TCB con TID <%d> y PID <%d>", tid, pid)
			continue
		}

		// Informar a memoria sobre la finalización del hilo
		request := types.RequestToMemory{
			Thread:    types.Thread{PID: execTCB.FatherPCB.PID, TID: execTCB.TID},
			Type:      types.FinishThread,
			Arguments: []string{},
		}
		err = sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error en la request de memoria sobre la finalización del hilo - %v", err)
		}

		// Desbloquear hilos que estaban bloqueados esperando el término de este TID
		moveBlockedThreadsByJoin(tid)

		// Liberar los mutexes que tenía el hilo que se está finalizando
		releaseMutexes(tid)

		// Limpiar el ExecStateThread si el hilo finalizado era el ejecutándose
		if kernelglobals.ExecStateThread != nil && kernelglobals.ExecStateThread.TID == types.Tid(tid) {
			kernelglobals.ExecStateThread = nil
		}

		kernelsync.SemMovedFinishThreads <- struct{}{}

		logger.Info("## Finalización del TID <%d> del PCB con PID <%d> completada", tid, pid)
	}
}

func moveBlockedThreadsByJoin(tidFinalizado int) {
	// Obtener el tamaño inicial de la cola de bloqueados
	blockedQueueSize := kernelglobals.BlockedStateQueue.Size()
	for i := 0; i < blockedQueueSize; i++ {
		// Obtener y remover el siguiente TCB de la cola de bloqueados
		tcb, err := kernelglobals.BlockedStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error al obtener el siguiente TCB de BlockedStateQueue: %v", err)
			continue
		}

		// Verificar que el campo JoinedTCB no sea nil antes de acceder a su TID
		if tcb.JoinedTCB != nil && tcb.JoinedTCB.TID == types.Tid(tidFinalizado) {
			tcb.JoinedTCB = nil // Resetear el campo JoinedTCB

			// Agregar el hilo a la cola de Ready
			err = kernelglobals.ShortTermScheduler.AddToReady(tcb)
			if err != nil {
				logger.Error("Error al agregar el TID <%d> del PCB con PID <%d> a la cola de Ready: %v", tcb.TID, tcb.FatherPCB.PID, err)
			} else {
				logger.Info("## Moviendo el TID <%d> del PCB con PID <%d> de estado BLOCK a estado READY por THREAD_JOIN", tcb.TID, tcb.FatherPCB.PID)
			}
		} else {
			// Si el hilo no estaba esperando, volver a agregarlo a la cola de bloqueados
			kernelglobals.BlockedStateQueue.Add(tcb)
		}
	}
}

func releaseMutexes(tid int) {
	execTCB := kernelglobals.ExecStateThread
	if execTCB == nil {
		logger.Error("No hay hilo en ejecución para liberar mutexes.")
		return
	}

	pcb := execTCB.FatherPCB

	for i, mutex := range pcb.CreatedMutexes {
		if mutex.AssignedTCB != nil && mutex.AssignedTCB.TID == types.Tid(tid) {
			// Liberar el mutex y desbloquear el primer hilo bloqueado
			mutex.AssignedTCB = nil // Liberar el mutex
			logger.Info("## Liberando el mutex <%s> del TID <%d>", mutex.Name, tid)

			execTCB.LockedMutexes = nil

			// Si hay hilos bloqueados en el mutex, mover el primero a Ready
			if len(mutex.BlockedTCBs) > 0 {
				nextThread := mutex.BlockedTCBs[0]
				mutex.BlockedTCBs = mutex.BlockedTCBs[1:] // Remover el primer hilo bloqueado de la lista

				// Asignar el mutex al siguiente hilo
				mutex.AssignedTCB = nextThread
				nextThread.LockedMutexes = append(nextThread.LockedMutexes, &mutex)

				kernelglobals.BlockedStateQueue.Remove(nextThread)

				// Mover el hilo a la cola de Ready
				err := kernelglobals.ShortTermScheduler.AddToReady(nextThread)
				if err != nil {
					logger.Error("Error al mover el TID <%d> del PCB con PID <%d> de estado BLOCK a READY: %v", nextThread.TID, nextThread.FatherPCB.PID, err)
				} else {
					logger.Info("## Asignando el mutex <%s> al TID <%d> del PCB con PID <%d> y moviendo a estado READY", mutex.Name, nextThread.TID, nextThread.FatherPCB.PID)
				}
			} else {
				logger.Info("## No hay hilos bloqueados esperando el mutex <%s>. Se ha liberado.", mutex.Name)
			}

			// Actualizar el mutex en el PCB
			pcb.CreatedMutexes[i] = mutex
		}
	}
}

func sendMemoryRequest(request types.RequestToMemory) error {
	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")

	// Serializar mensaje
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	logger.Info("CONTENIDO DE REQUEST A MEMORIA:\n	-Type: %v\n	-Arguments: %v", request.Type, request.Arguments)

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/%s", kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort, request.Type)
	logger.Debug("Enviando request a memoria")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Recibo repuesta de memoria
	resp, err := memoria.Do(req)
	if err != nil {
		return err
	}

	err = handleMemoryResponseError(resp, request.Type)
	if err != nil {
		return err
	}
	return nil
}

// esta funcion es auxiliar de sendMemoryRequest
func handleMemoryResponseError(response *http.Response, TypeRequest string) error {
	if response.StatusCode != http.StatusOK {
		err := types.ErrorRequestType[TypeRequest]
		return err
	}
	return nil
}
