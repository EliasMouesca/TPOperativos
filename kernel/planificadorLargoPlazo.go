package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

func planificadorLargoPlazo() {
	// En el enunciado en implementacion dice que hay que inicializar un proceso
	// quiza hay que hacerlo aca o en kernel.go es lo mismo creo
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

		// Crear el request para verificar si memoria tiene espacio
		request := types.RequestToMemory{
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
		pidStr := strconv.Itoa(int(PID))
		request := types.RequestToMemory{
			Type:      types.FinishProcess,
			Arguments: []string{pidStr},
		}
		logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", PID)

		// Simular el request a memoria (puede reemplazarse con la función real si está disponible)
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error en el request a memoria: %v", err)
		}
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
		//kernelsync.SemThreadCreate <- 0
	}
}

func ThreadToExit() {
	// Al momento de finalizar un hilo, el Kernel deberá informar a la Memoria
	// la finalización del mismo y deberá mover al estado READY a todos los
	// hilos que se encontraban bloqueados por ese TID. De esta manera, se
	// desbloquean aquellos hilos bloqueados por THREAD_JOIN y por mutex
	// tomados por el hilo finalizado (en caso que hubiera).
	execTCB := kernelglobals.ExecStateThread
	currentPCB := execTCB.FatherPCB
	for {
		TID := <-kernelsync.ChannelFinishThread
		tid := strconv.Itoa(TID)
		request := types.RequestToMemory{
			Type:      types.FinishThread,
			Arguments: []string{tid},
		}

		logger.Info("## Iniciando finalización del TID <%v> del PCB con PID <%d>", tid, currentPCB.PID)

		logger.Debug("Informando a Memoria sobre la finalización del hilo con TID %v", tid)
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error en la request de memoria sobre la finalizacion del hilo - %v", err)
		}

		// Desbloquear hilos que estaban bloqueados esperando el término de este TID
		//moveBlockedThreadsByJoin(TID)

		// Liberar los mutexes que tenía el hilo que se está finalizando
		//releaseMutexes(TID)

		// Mover el hilo actual a ExitStateQueue
		kernelglobals.ExitStateQueue.Add(execTCB)
		logger.Info("## Moviendo el TID <%d> al estado EXIT", TID)

		// Limpiar el ExecStateThread para indicar que no hay hilo en ejecución
		kernelglobals.ExecStateThread = nil

		logger.Info("## Finalización del TID <%d> del PCB con PID <%d> completada", TID, currentPCB.PID)
	}
}

/*
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

		// Si el hilo estaba esperando al tidFinalizado, moverlo a la cola de Ready
		if tcb.JoinedTCB == tidFinalizado {
			tcb.JoinedTCB = -1 // Resetear el campo JoinedTCB

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
	tcb := kernelglobals.ExecStateThread
	for _, mutexID := range tcb.LockedMutexes {
		mutexWrapper, exists := kernelglobals.GlobalMutexRegistry[mutexID]
		if !exists {
			logger.Error("## No se encontró el mutex con ID <%d> en el registro global", mutexID)
			continue
		}

		mutexWrapper.Mutex.Lock()
		mutexWrapper.AssignedTID = -1 // Marcar el mutex como libre
		logger.Info("## Liberando el mutex <%d> del TID <%d>", mutexID, tcb.TID)

		if len(mutexWrapper.BlockedTCBs) > 0 {
			nextThread := mutexWrapper.BlockedTCBs[0]
			mutexWrapper.BlockedTCBs = mutexWrapper.BlockedTCBs[1:]
			mutexWrapper.AssignedTID = nextThread.TID
			nextThread.Mutex = append(nextThread.Mutex, mutexID)
			err := kernelglobals.ShortTermScheduler.AddToReady(nextThread)
			if err != nil {
				logger.Error("Error al mover el TID <%d> del PCB con PID <%d> de estado BLOCK a READY: %v", nextThread.TID, nextThread.FatherPCB.PID, err)
			} else {
				logger.Info("## Asignando el mutex <%d> al TID <%d> del PCB con PID <%d> y moviendo a estado READY", mutexID, nextThread.TID, nextThread.FatherPCB.PID)
			}
		} else {
			logger.Info("## No hay hilos bloqueados esperando el mutex <%d>. Se ha liberado.", mutexID)
		}

		mutexWrapper.Mutex.Unlock()
	}
}*/

func sendMemoryRequest(request types.RequestToMemory) error {
	// Simulación de respuesta exitosa sin hacer la solicitud real a memoria
	logger.Debug("Simulando respuesta exitosa de la memoria para request de tipo %s", request.Type)
	return nil
}

/*
func sendMemoryRequest(request types.RequestToMemory) error {
	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")

	// Serializar mensaje
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/"+request.Type, kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort)
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
*/
