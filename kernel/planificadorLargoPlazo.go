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

	//DUMP MEMORY
	case 0:

	// IO
	case 1:

	//PROCESS_CREATE
	case 2:
		pseudocodigo := syscall.Arguments[0]
		processSize, _ := strconv.Atoi(syscall.Arguments[1])
		prioridad, _ := strconv.Atoi(syscall.Arguments[2])

		if len(NEW) != 0 {
			NEW = append(NEW, types.PCB{})
			// Si NEW no esta vacio significa que hay un proceso esperando a ser mandado a Ready
			// Habria que hacer una sincronizacion de los procesos que vayan llegando
			// y que vayan preguntando a memoria a medida que pasan los procesos a Ready
		}

		PROCESS_CREATE(pseudocodigo, processSize, prioridad)
		logger.Info("## (<PID>:0) Se crea el proceso - Estado: NEW")
		for available == 0 {
			go availableMemory(processSize)
			if available == 1 {
				// Aca habria que ver si usamos Thread_Create o harcodeado
			}
		}

	//CREATE_THREAD
	case 3:

	//THREAD_JOIN
	case 4:

	//THREAD_CANCEL
	case 5:

	//MUTEX_CREATE
	case 6:

	// MUTEX_LOCK
	case 7:

	//MUTEX_UNLOCK
	case 8:

		//THREAD_EXIT
	case 9:

		//PROCESS_EXIT
	case 10:

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
