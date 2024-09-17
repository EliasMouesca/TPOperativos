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
		processToReady()
	}()

	kernelsync.WaitPlanificadorLP.Wait()
}

func processToReady() {
	for {
		// Espera a que se cree un proceso y le mande sus argumentos,
		// se van guardando los argumentos de cada proceso en el canal a medidad que se crean
		args := <-kernelsync.ChannelProcessArguments
		fileName := args[0]
		processSize, _ := strconv.Atoi(args[1])
		prioridad, _ := strconv.Atoi(args[2])

		// Se crea un hilo porque tiene que esperar a que se libere espacio en memoria
		// para mandar el siguiente proceso a Ready
		kernelsync.WaitPlanificadorLP.Add(1)
		go func() { // lo testie y funciona, con esto podemos hacer un availableMemory polimorfico, quiza, esta en prueba todavia
			defer kernelsync.WaitPlanificadorLP.Done()
			availableProcessMemory(processSize, fileName)
		}()

		<-kernelsync.ChannelMemoryRequest
		// Se libero espacio en memoria

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
	kernelsync.MutexPlanificadorLP.Unlock()

	logger.Debug("Informando a Memoria sobre la finalización del proceso con PID %d", pcb.PID)
	//informarMemoriaProcessToExit(pcb.PID)
}

// esto hay que mejorarlo seguro quiza hacerlo de alguna manera
// polimorfica, ya que lo unico que hace  basicamente
// largo plazo es comunicarse con memoria
func availableProcessMemory(processSize int, fileName string) {
	for {
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
			logger.Fatal("Error al serializar request - %v", err)
			return
		}

		// Hacer request a memoria
		memoria := &http.Client{}
		url := fmt.Sprintf("http://%s:%d/memoria/availableMemory", kernelglobals.Config.MemoryAddress, kernelglobals.Config.MemoryPort)
		logger.Debug("Enviando request a memoria")
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(request_json))
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
			kernelsync.ChannelMemoryRequest <- 0
		} else {
			logger.Info("No hay espacio disponible en memoria")
			return
		}
	}
}
