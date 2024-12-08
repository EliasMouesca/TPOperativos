package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
	"time"
)

func planificadorLargoPlazo() {
	// En el enunciado en implementacion dice que hay que inicializar un proceso
	// quiza hay que hacerlo aca o en kernel.go es lo mismo creo

	logger.Info("Iniciando el planificador de largo plazo")

	go NewProcessToReady()
	go ProcessToExit()
	go NewThreadToReady()
	go ThreadToExit()
	go UnlockIO()
}

var puedeSolicitar int

func NewProcessToReady() {
	for {
		// Espera los argumentos del proceso desde el canal
		args := <-kernelsync.ChannelProcessArguments
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
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Warn("Error al enviar request a memoria: %v", err)
			if errors.Is(err, errors.New("memoria: Se debe compactar")) {
				solicitarCompactacion(request, prioridad)
			} else {
				go func(request types.RequestToMemory, pid int, prioridad int) {
					for {
						logger.Info("<PID: %v> Esperando a que finalice un proceso para crearse", pid)
						<-kernelsync.InitProcess // Espera a que finalice otro proceso antes de intentar de nuevo
						logger.Debug("Termino un proceso, intentando crear proceso pendiente: %v", pid)
						err1 := sendMemoryRequest(request)
						if err1 == nil {
							logger.Debug("Se librero un proceso y ahora hay espacio para crear el proceso con PID: %v", request.Thread.PID)
							logger.Info("Creando proceso pendiente (PID: %v)", pid)
							agregarProcesoAReady(pid, prioridad)
							kernelsync.MutexPuedoCrearProceso.Unlock()
							break
						} else {
							// Si entra aca y no al if este, es que no hay espacio dispoible ni compactando entocnes sigue
							// esperando a que se libere otro proceso
							if errors.Is(err1, types.ErrorRequestType[types.Compactacion]) {
								solicitarCompactacion(request, prioridad)
								kernelsync.MutexPuedoCrearProceso.Unlock()
								break
							}
						}
					}
				}(request, pid, prioridad)
			}
		} else {
			logger.Debug("Hay espacio disponible en memoria")
			agregarProcesoAReady(pid, prioridad)
			kernelsync.MutexPuedoCrearProceso.Unlock()
		}
	}
}

func solicitarCompactacion(request types.RequestToMemory, prioridad int) {
	logger.Debug("Memoria necesita compactar para crear el proceso...")
	requestCompact := types.RequestToMemory{
		Thread:    request.Thread,
		Type:      types.Compactacion,
		Arguments: []string{},
	}

	kernelsync.MutexCPU.Lock()

	errCompact := sendMemoryRequest(requestCompact)
	if errCompact != nil {
		logger.Error("Error al enviar request de compactar a memoria: %v", errCompact)
		kernelsync.MutexCPU.Unlock()
		return
	}
	kernelsync.MutexCPU.Unlock()
	logger.Debug("Se compacto, entonces mandamos a crear el proceso otra vez")

	errNewRequest := sendMemoryRequest(request)
	if errNewRequest != nil {
		logger.Error("Error al enviar request a memoria: %v", errNewRequest)
		return
	}
	agregarProcesoAReady(int(request.Thread.PID), prioridad)
}

func agregarProcesoAReady(pid int, prioridad int) {
	logger.Debug("Entra a crear proceso aux el PID: %v", pid)
	pcbPtr := buscarPCBPorPID(types.Pid(pid))
	if pcbPtr == nil {
		logger.Error("No se encontró el PCB con PID <%d> en la lista global", pid)
	}

	// Crear el hilo principal (mainThread) ahora que el proceso tiene espacio en memoria
	mainThread := kerneltypes.TCB{
		TID:       0,
		Prioridad: prioridad,
		FatherPCB: pcbPtr,
	}

	// Agregar el mainThread a la lista de TCBs en el kernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread)
	logger.Debug("Se agrega a EVERYTCB el TCB con TID: %v del Proceso con PID: %v", mainThread.TID, pid)
	// Obtener el puntero del hilo principal para encolarlo en Ready
	mainThreadPtr := buscarTCBPorTID(0, pcbPtr.PID)

	// Mover el mainThread a la cola de Ready
	kernelglobals.ShortTermScheduler.AddToReady(mainThreadPtr)
	logger.Info("## (<%d:0>) Se movio a la cola Ready", pid)

	// Señalización para indicar que el proceso ha sido agregado exitosamente a Ready
	kernelsync.SemProcessCreateOK <- struct{}{}
}

func ProcessToExit() {
	for {
		// Recibir la señal de finalización de un proceso
		PID := <-kernelsync.ChannelFinishprocess
		logger.Debug("entra a process to exit despues del channel")
		request := types.RequestToMemory{
			Thread:    types.Thread{PID: PID},
			Type:      types.FinishProcess,
			Arguments: []string{},
		}

		logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", PID)
		err := sendMemoryRequest(request)
		if err != nil {
			logger.Error("Error al terminar un proceso: %v", err)
			return
		}

		logger.Debug("Se informo a memoria correctamente")
		go func() {
			kernelsync.InitProcess <- struct{}{}
		}()
		logger.Debug("*** Termino un ciclo de process to exit ***")
		kernelsync.ChannelFinishProcess2 <- true
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
		logger.Info("## (<%d>:<%d>) Se crea el Hilo- Estado: READY", newTCB.FatherPCB.PID, newTCB.TID)

		kernelsync.ThreadCreateComplete <- struct{}{}

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

		logger.Trace("Se llamó thread to exit TID <%v> del PCB con PID <%d>", tid, pid)

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
		moveBlockedThreadsByJoin(tid, pid)

		// Liberar los mutexes que tenía el hilo que se está finalizando
		releaseMutexes(tid)

		// Limpiar el ExecStateThread si el hilo finalizado era el ejecutándose
		if kernelglobals.ExecStateThread != nil && kernelglobals.ExecStateThread.TID == types.Tid(tid) {
			kernelglobals.ExecStateThread = nil
		}

		kernelsync.ThreadExitComplete <- struct{}{}
	}
}

func moveBlockedThreadsByJoin(tidFinalizado int, pidFinalizado int) {
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
		if tcb.JoinedTCB != nil && tcb.JoinedTCB.TID == types.Tid(tidFinalizado) &&
			tcb.FatherPCB.PID == types.Pid(pidFinalizado) {

			tcb.JoinedTCB = nil // Resetear el campo JoinedTCB

			// Agregar el hilo a la cola de Ready
			err = kernelglobals.ShortTermScheduler.AddToReady(tcb)
			if err != nil {
				logger.Error("Error al agregar el TID <%d> del PCB con PID <%d> a la cola de Ready: %v", tcb.TID, tcb.FatherPCB.PID, err)
			} else {
				logger.Info("## (<%v>:<%v>) Moviendo de estado BLOCK a estado READY por THREAD_JOIN", tcb.FatherPCB.PID, tcb.TID)
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

func UnlockIO() {
	for {
		tcbBlock := <-kernelsync.ChannelIO
		timeBlocked := <-kernelsync.ChannelIO2

		time.Sleep(time.Duration(timeBlocked) * time.Millisecond)
		err := kernelglobals.BlockedStateQueue.Remove(tcbBlock)

		if err != nil {
			logger.Error("No se pudo remover el tcb de la BlockQueue - %v", err)
		}

		err = kernelglobals.ShortTermScheduler.AddToReady(tcbBlock)
		if err != nil {
			logger.Error("No se pudo mover el tcb a la cola Ready. - %v", err)
		}

		//logger.Debug("-- Probando cosas raras --")

		//kernelsync.SyscallFinalizada <- true
		//if kernelglobals.Config.SchedulerAlgorithm == "CMN" {
		//	// Termino de ejecutar la Syscall => Reinicia el Quantum
		//	go func() {
		//		logger.Warn("Reiniciamos timer por syscall")
		//		kernelsync.SyscallChannel <- struct{}{}
		//	}()
		//}

		logger.Info("“## (<%v>:<%v>) finalizó IO y pasa a READY", tcbBlock.FatherPCB.PID, tcbBlock.TID)
	}
}

func sendMemoryRequest(request types.RequestToMemory) error {
	logger.Debug("Enviando request a  memoria: %v para el THREAD: %v", request.Type, request.Thread)

	// Serializar mensaje
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Hacer request a memoria
	url := fmt.Sprintf("http://%s:%d/memoria/%s", kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort, request.Type)
	logger.Debug("Enviando request a memoria: %v", url)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		logger.Error("Error al realizar request memoria: %v", err)
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
	logger.Debug("Memoria respondio a: %v con: %v", TypeRequest, response.StatusCode)
	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusConflict { // Conflict es compactacion.
			err := types.ErrorRequestType[types.Compactacion]
			return err
		}
		err := types.ErrorRequestType[TypeRequest]
		return err
	}
	return nil
}
