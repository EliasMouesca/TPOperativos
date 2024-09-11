package main

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

var PIDcount int = 0

func PROCESS_CREATE(fileName string, processSize int, prioridad int) {
	// Se crea el PCB y el Hilo 0
	var procesoCreado types.PCB
	PIDcount++
	procesoCreado.PID = PIDcount

	hiloMain := types.TCB{TID: 0, Prioridad: prioridad}
	procesoCreado.TIDs = []types.TCB{hiloMain}

	logger.Info("## (<%P>:0) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	//// Se agrega el proceso a NEW
	NEW = append(NEW, procesoCreado)
}

func THREAD_CREATE(pseudocodigo string, prioridad int) {
	var hiloCreado types.TCB
	hiloCreado.TID = len(Ready)
	hiloCreado.Prioridad = prioridad

	Ready = append(Ready, hiloCreado)
}
