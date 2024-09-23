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
