package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func planificadorLargoPlazo() {
	go processToReady()
	go processToExit()
}

func processToReady() {
	var available int = 0
	for {
		args := <-kernelsync.ChannelProcessArguments
		fileName := args[0]
		processSize, _ := strconv.Atoi(args[1])
		prioridad, _ := strconv.Atoi(args[2])

		availableMemory(processSize, fileName)
		if !kernelglobals.NewStateQueue.IsEmpty() && available == 1 {
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
			kernelglobals.ReadyStateQueue.Add(&mainThread)
			available = 0 // reiniciar available
		}
	}
}

func processToExit() {

	tcb := kernelglobals.ExecStateThread
	pcb := tcb.ConectPCB

	queueSize := kernelglobals.ReadyStateQueue.Size()
	for i := 0; i < queueSize; i++ {
		readyTCB, err := kernelglobals.ReadyStateQueue.GetAndRemoveNext()
		if err != nil {
			logger.Error("Error al obtener el siguiente TCB de ReadyStateQueue - %v", err)
			continue
		}

		// Verificar si el TCB pertenece al mismo PCB que el proceso que está finalizando
		if readyTCB.ConectPCB == pcb {
			// Si el TCB pertenece al PCB, lo eliminamos de la cola y no lo reinsertamos
			logger.Debug("Eliminando TCB con TID %d del proceso con PID %d de ReadyStateQueue", readyTCB.TID, pcb.PID)
		} else {
			// Si no pertenece, lo volvemos a insertar en la cola
			kernelglobals.ReadyStateQueue.Add(readyTCB)
		}
	}
	kernelsync.PlanificadorLPMutex.Unlock()

	logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", pcb.PID)
	//informarMemoriaProcessToExit(pcb.PID)
	processToReady()

	// aca hay que hacer sincronizacion
	// por que hay q informar a memoria
	// y despues volver al flujo de PROCESS_EXIT
	// y despues volver al flujo de esta funcion con processToReady
	// vamos por buen camino processToReady se tiene que inicializar por el enunciado
	// hay que ver el tema de sincroo!!
}

// esto hay que mejorarlo seguro quiza hacerlo de alguna manera
// polimorfica, ya que lo unico que hace  basicamente
// largo plazo es comunicarse con memoria
func availableMemory(processSize int, fileName string) {

	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")
	request := struct {
		ProcessSize int
		FileName    string
	}{
		ProcessSize: processSize,
		FileName:    fileName,
	}

	// Serializar mensaje
	request_json, err := json.Marshal(request)
	if err != nil {
		logger.Fatal("Error al serializar processSize - %v", err)
		return
	}

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/availableMemory", kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort)
	logger.Debug("Enviando request a memoria")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(processSize_json))
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
	if resp.StatusCode != http.StatusOK {
		available = 1
	} else {
		logger.Info("No hay espacio disponible en memoria")
	}
}
