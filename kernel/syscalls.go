package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

var PIDcount int = 0
var available int = 0

func PROCESS_CREATE(fileName string, processSize int, prioridad int) {
	// Se crea el PCB
	var procesoCreado types.PCB
	PIDcount++
	procesoCreado.PID = PIDcount

	// Creamos TCB asociado al proceso con TID 0 y prioridad como argumento que nos pasa CPU
	hiloMain := types.TCB{TID: 0, Prioridad: prioridad}
	procesoCreado.TIDs = []*types.TCB{&hiloMain}

	logger.Info("## (<%T>:%v) Se crea el proceso - Estado: NEW", procesoCreado.PID, procesoCreado.TIDs[hiloMain.TID])
	// si es el hiloMain es 0 te va a mostrar el hilo posicionado en el indice 0, que deberia ser el hiloMain :)

	// Se agrega el proceso a NEW
	NEW = append(NEW, procesoCreado)

	// Preguntar a memoria si hay espacio para pasarlo a Ready
	logger.Debug("Preguntando a memoria si tiene espacio disponible")
	for available == 0 {
		go availableMemory(processSize)
		if available == 1 {
			logger.Debug("Se a liberado espacio en memoria para el proceso")
			procesoCreado := NEW[0]
			NEW = NEW[1:]
			logger.Debug("Trasladar proceso de NEW a Ready")
			Ready = append(Ready, procesoCreado)
		}
	}
}

func availableMemory(processSize int) {
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
