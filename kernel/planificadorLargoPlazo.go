package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

var available int = 0
var NEW []types.PCB
var Ready []types.TCB

func planificadorLargoPlazo(syscall syscalls.Syscall) {
	logger.Info("## (<PID>:<TID>) - Solicitó syscall: <%v>", syscall.Description)
	switch syscall.Type {

	//PROCESS_CREATE
	case 2:
		pseudocodigo := syscall.Arguments[0]
		processSize, _ := strconv.Atoi(syscall.Arguments[1])
		prioridad, _ := strconv.Atoi(syscall.Arguments[2])

		PROCESS_CREATE(pseudocodigo, processSize, prioridad)
		logger.Info("## (<PID>:0) Se crea el proceso - Estado: NEW")
		for available == 0 {
			go availableMemory(processSize)
			if available == 1 {
				// Aca habria que ver si usamos Thread_Create o harcodeado
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
