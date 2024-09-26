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

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		ProcessToReady()
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		ProcessToExit()
	}()

	kernelsync.WaitPlanificadorLP.Wait()
}

func ProcessToReady() {
	for {
		// Espera a que se cree un proceso y le mande sus argumentos,
		// se van guardando los argumentos de cada proceso en el canal a medidad que se crean
		args := <-kernelsync.ChannelProcessArguments
		logger.Debug("Llegaron los argumentos de la syscall")
		fileName := args[0]
		processSize := args[1]
		prioridad, _ := strconv.Atoi(args[2])

		request := types.RequestToMemory{
			Type:      types.CreateProcess,
			Arguments: []string{fileName, processSize},
		}
		// Se crea un hilo porque tiene que esperar a que se libere espacio en memoria
		// para mandar el siguiente proceso a Ready
		logger.Debug("Preguntando a memoria si tiene espacio disponible")
		kernelsync.WaitPlanificadorLP.Add(1)
		go func() {
			defer kernelsync.WaitPlanificadorLP.Done()
			for {
				err := sendMemoryRequest(request)
				if err != nil {
					logger.Debug("Error en la request de memoria sobre la creacion del proceso- %v", err)
					<-kernelsync.InitProcess
				} else {
					logger.Debug("Hay espacio disponible en memoria")
					break
				}
			}
		}()

		// El if para preguntar si esta vacia la cola Null no hace falta,
		// porque esta planificacion solo ocurre si se creo el proceso,
		// el cual se envian sus argumentos atraves de ChannelProcessArguments
		pcb, err := kernelglobals.NewStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error en la cola NEW - %v", err)
		}

		// busco al tcb con TID = 0 del pcb que se obtuvo de la cola de NEW
		var mainThread kerneltypes.TCB
		for _, tid := range pcb.TIDs {
			if tid == 0 {
				mainThread = kerneltypes.TCB{
					TID:       tid,
					Prioridad: prioridad,
					FatherPCB: pcb,
				}
				break
			}
		}

		// Mandamos el hiloMain a Ready
		kernelglobals.ShortTermScheduler.AddToReady(&mainThread)
		logger.Info("Se agrego el hilo main a la cola Ready")
	}
}

func ProcessToExit() {
	for {
		PID := <-kernelsync.ChannelFinishprocess
		pid := strconv.Itoa(PID)
		request := types.RequestToMemory{
			Type:      types.FinishProcess,
			Arguments: []string{pid},
		}
		logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", PID)
		kernelsync.Finishprocess.Add(1)
		go func() {
			defer kernelsync.Finishprocess.Done()
			for {
				err := sendMemoryRequest(request)
				if err != nil {
					logger.Error("Error en la request de memoria sobre la finalizacion del proceso - %v", err)
				} else {
					kernelsync.SemFinishprocess <- 0
				}
			}
		}()
	}
}

func ThreadToReady() {

}

func ThreadEnding() {
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

		logger.Info("## Iniciando finalización del TID <%d> del PCB con PID <%d>", tid, currentPCB.PID)

		logger.Debug("Informando a Memoria sobre la finalización del hilo con TID %d", tid)
		kernelsync.FinishThread.Add(1)
		go func() {
			defer kernelsync.FinishThread.Done()
			for {
				err := sendMemoryRequest(request)
				if err != nil {
					logger.Error("Error en la request de memoria sobre la finalizacion del hilo - %v", err)
				} else {
					kernelsync.SemFinishThread <- 0
				}
			}
		}()

		// Desbloquear hilos que estaban bloqueados esperando el término de este TID
		moveBlockedThreadsByJoin(TID)

		// Liberar los mutexes que tenía el hilo que se está finalizando
		releaseMutexes(TID)

		// Mover el hilo actual a ExitStateQueue
		kernelglobals.ExitStateQueue.Add(execTCB)
		logger.Info("## Moviendo el TID <%d> al estado EXIT", TID)

		// Limpiar el ExecStateThread para indicar que no hay hilo en ejecución
		kernelglobals.ExecStateThread = nil

		logger.Info("## Finalización del TID <%d> del PCB con PID <%d> completada", TID, currentPCB.PID)
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
}

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

	err = handleMemoryResponse(resp, request.Type)
	if err != nil {
		return err
	}
	return nil
}

// esta funcion es auxiliar de sendMemoryRequest
func handleMemoryResponse(response *http.Response, TypeRequest string) error {
	if response.StatusCode != http.StatusOK {
		err := types.ErrorRequestType[TypeRequest]
		return err
	}
	return nil
}
