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

type syscallFunc func(args ...interface{})

var syscallSet = map[string]syscallFunc{
	"PROCESS_CREATE": PROCESS_CREATE,
	"PROCESS_EXIT":   PROCESS_EXIT,
	"THREAD_CREATE":  THREAD_CREATE,
	// "THREAD_JOIN": THREAD_JOIN,
	// "THREAD_CANCEL": THREAD_CANCEL
	// "THREAD_EXIT": THREAD_CREATE,
	// "MUTEX_CREATE": MUTEX_CREATE,
	// "MUTEX_LOCK": MUTEX_LOCK,
	// "MUTEX_UNLOCK": MUTEX_UNLOCK,
}

var PIDcount int = 0

func PROCESS_CREATE(args ...interface{}) {
	pseudoCodigo := args[0]
	processSize := args[1].(int)
	prioridad := args[2].(int)

	// Se crea el PCB y el Hilo 0

	var procesoCreado types.PCB
	PIDcount++
	procesoCreado.PID = PIDcount
	procesoCreado.TIDs = []types.TCB{hiloMain}
	hiloMain := types.TCB{
		TID:       0,
		Prioridad: prioridad,
	}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	// Se agrega el proceso a NEW
	global.NEW = append(global.NEW, procesoCreado)
}

func PROCESS_EXIT(args ...interface{}) {
	fmt.Println("Proceso finalizado")
}

func THREAD_CREATE(args ...interface{}) {
	fmt.Println("Creando hilo...")
}

func THREAD_JOIN(args ...interface{}) {
	fmt.Println("Esperando a que el hilo termine...")
}

func THREAD_CANCEL(args ...interface{}) {
	fmt.Println("Cancelando hilo...")
}

func THREAD_EXIT(args ...interface{}) {
	fmt.Println("Saliendo del hilo...")
}

func MUTEX_CREATE(args ...interface{}) {
	fmt.Println("Creando mutex...")
}

func MUTEX_LOCK(args ...interface{}) {
	fmt.Println("Bloqueando mutex...")
}

func MUTEX_UNLOCK(args ...interface{}) {
	fmt.Println("Desbloqueando mutex...")
}

func ExecuteSyscall(syscallName string, args ...interface{}) {
	if syscallFunc, exists := syscallSet[syscallName]; exists {
		syscallFunc(args...)
	} else {
		fmt.Println("Syscall no encontrada:", syscallName)
	}
}

//ALGUNAS FUNCIONES AUXILIARES

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
