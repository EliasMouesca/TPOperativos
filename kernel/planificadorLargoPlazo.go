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
		processToReady()
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToExit()
	}()

	kernelsync.WaitPlanificadorLP.Wait()
}

func processToReady() {
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
		kernelsync.InitProcess.Add(1)
		go func() {
			defer kernelsync.InitProcess.Done()
			sendMemoryRequest(request)
		}()

		// Memory Semaphore Create Process
		<-kernelsync.SemCreateprocess
		// Se libero espacio en memoria.

		// El if para preguntar si esta vacia la cola New no hace falta,
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
					ConectPCB: pcb,
				}
				break
			}
		}

		// Mandamos el hiloMain a Ready
		kernelglobals.ReadyStateQueue.Add(&mainThread)
		logger.Info("Se agrego el hilo main a la cola Ready")
	}
}

func processToExit() {
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
			sendMemoryRequest(request)
		}()
	}
}

func sendMemoryRequest(request types.RequestToMemory) {
	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")

	// Serializar mensaje
	jsonRequest, err := json.Marshal(request)
	if err != nil {
		logger.Fatal("Error al serializar request - %v", err)
		return
	}

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/"+request.Type, kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort)
	logger.Debug("Enviando request a memoria")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
	if err != nil {
		logger.Fatal("Error al conectar con memoria - %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Recibo repuesta de memoria
	resp, err := memoria.Do(req)
	if err != nil {
		logger.Fatal("Error al obtener mensaje de respuesta por parte de memoria - %v", err)
		return
	}

	err = handleMemoryResponse(resp, request.Type)
	if err != nil {
		logger.Error("Memoria respondio con un error - %v", err)
	}
}

// esta funcion es auxiliar de sendMemoryRequest
func handleMemoryResponse(response *http.Response, TypeRequest string) error {
	if response.StatusCode != http.StatusOK {
		err := types.ErrorRequestType[TypeRequest]
		return err
	}

	switch TypeRequest {
	case types.CreateProcess:
		kernelsync.SemCreateprocess <- 0
	case types.FinishProcess:
		kernelsync.SemFinishprocess <- 0
	case types.CreateThread:

	case types.FinishThread:

	case types.MemoryDump:

	}
	return nil
}
