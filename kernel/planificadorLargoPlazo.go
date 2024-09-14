package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

// nose todavia bien como va a ser el planificador
// si poner estas funciones dentro de las syscalls
// o crear un hilo en Kernel.go que quede corriendo

var available int = 0

func processToReady(processSize int, prioridad int) {
	for true {
		availableMemory(processSize) // seguro hay qui enviarle el psudoCodigo a memoria
		if !global.NEW.IsEmpty() && available == 1 {
			pcb := global.NEW.GetAndRemoveNext()
			hiloMain := types.TCB{
				TID:       0,
				Prioridad: prioridad,
			}
			available = 0 // reiniciar available
			// algorithm.AddToReady(hiloMain)
		}
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
	url := fmt.Sprintf("http://%s:%d/memoria/availableMemory", Config.MemoryAddress, Config.MemoryPort)
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
