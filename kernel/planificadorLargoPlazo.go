package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

var available int = 0
var NEW []types.PCB
var Ready []types.TCB

func planificadorLargoPlazo(args ...interface{}) {
	syscallName := args[0].(string)

	switch syscallName {

	//PROCESS
	case "PROCESS_CREATE":
		pseudocodigo := args[1].(string)
		processSize := args[2].(int)
		prioridad := args[3].(int)

		for available == 0 {
			go availableMemory(processSize)
			if available == 1 {
				PROCESS_CREATE(pseudocodigo, processSize, prioridad)
			}
		}
	case "PROCESS_EXIT":

	//THREADS
	case "CREATE_THREAD":
	case "THREAD_JOIN":
	case "THREAD_CANCEL":
	case "THREAD_EXIT":

	//MUTEX
	case "MUTEX_CREATE":
	case "MUTEX_LOCK":
	case "MUTEX_UNLOCK":

	//MEMORY
	case "DUMP_MEMORY":

	}
}

func availableMemory(processSize int) {

	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")

	// Serializar mensaje
	processSize_json, err := json.Marshal(processSize)
	if err != nil {
		logger.Fatal("Error al serializar processSize - %v", err)
		return
	}

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/availableMemory", config.MemoryAddress, config.MemoryPort)
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
