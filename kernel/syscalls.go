package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

func PROCESS_CREATE(fileName string, processSize int, TID int) {
	// Crear proceso
	var procesoCreado PCB
	procesoCreado.PID = 1

	// Pregunta a memoria si hay espacio
	processSize_json, err := json.Marshal(processSize)
	if err != nil {
		return
	}

	//Evaluar error
	cliente := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/accion", config.MemoryAddress, config.MemoryPort)
	logger.Debug("Enviando request a %v", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(processSize_json))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cliente.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	respBody := string(respBytes)
	// Si hay espacio crea el TCB con TID en 0 y lo manda a Ready
	if respBody == "available" {
		hiloMain := types.TCB{TID: TID}
		procesoCreado.TIDs = []*types.TCB{&hiloMain}
		Ready = append(Ready, procesoCreado)
	} else {
		// Si no hay espacio lo guarda en la cola de NEW
		NEW = append(NEW, procesoCreado)
	}
}
