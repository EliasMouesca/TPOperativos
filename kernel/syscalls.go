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

func PROCESS_CREATE(fileName string, processSize int, prioridad int) {
	// Se crea el PCB y el Hilo 0
	var procesoCreado types.PCB
	PIDcount++
	procesoCreado.PID = PIDcount

	hiloMain := types.TCB{PID: procesoCreado.PID,
		TID:       0,
		Prioridad: prioridad}

	procesoCreado.TIDs = []types.TCB{hiloMain}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	// Se agrega el proceso a NEW
	NEW = append(NEW, procesoCreado)

	// Mover el proceso a la cola READY si hay memoria disponible
	for available == 1 { // LUEGO DE ARREGLAR LO DE MEMORIA PONER EN  available = 0
		//go availableMemory(processSize)	COMENTADO PARA HACER TEST
		if available == 1 {
			Ready = append(Ready, hiloMain)
			logger.Info("## (%d:0) Proceso movido a READY", procesoCreado.PID)
		}
	}
}

func PROCESS_EXIT(pcb types.PCB) {
	logger.Info("## Finaliza el proceso <%d>", pcb.PID)

	// Liberar recursos de los TCBs del proceso
	for i := len(Ready) - 1; i >= 0; i-- {
		if Ready[i].PID == pcb.PID {
			Ready = append(Ready[:i], Ready[i+1:]...)
		}
	}

	//Deberia liberar memoria de este pcb
	//liberarMemoria(pcb.PID)

	// Remover el PCB de la cola NEW
	for i := len(NEW) - 1; i >= 0; i-- {
		if NEW[i].PID == pcb.PID {
			NEW = append(NEW[:i], NEW[i+1:]...)
			break
		}
	}
}

func THREAD_CREATE(pcb *types.PCB, pseudocodigo string, prioridad int) {
	// Crear el TCB para el nuevo hilo
	nuevoTID := len(pcb.TIDs)
	hiloCreado := types.TCB{
		PID:       pcb.PID,
		TID:       nuevoTID,
		Prioridad: prioridad,
	}

	// Agregar el nuevo TCB al PCB
	pcb.TIDs = append(pcb.TIDs, hiloCreado)

	// Mover el nuevo hilo a la cola READY
	Ready = append(Ready, hiloCreado)

	logger.Info("## (%d:%d) Se crea el hilo - Estado: READY", pcb.PID, hiloCreado.TID)
}

func THREAD_EXIT(tcb types.TCB) {
	logger.Info("## (%d:%d) Finaliza el hilo", tcb.PID, tcb.TID)

	// Remover el TCB de la cola READY
	for i := len(Ready) - 1; i >= 0; i-- {
		if Ready[i].PID == tcb.PID && Ready[i].TID == tcb.TID {
			Ready = append(Ready[:i], Ready[i+1:]...)
			break
		}
	}

	// Si era el último hilo, finalizar el proceso
	if len(getPCBByPID(tcb.PID).TIDs) == 1 {
		PROCESS_EXIT(getPCBByPID(tcb.PID))
	}
}

//ALGUNAS FUNCIONES AUXILIARES

func getPCBByPID(pid int) types.PCB {
	for _, pcb := range NEW {
		if pcb.PID == pid {
			return pcb
		}
	}
	return types.PCB{}
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
