package main

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"sync"
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

var mutex sync.Mutex
var PIDcount int = 0

func PROCESS_CREATE(args ...interface{}) {
	pseudoCodigo := args[0]
	processSize := args[1].(int)
	prioridad := args[2].(int)

	// Se crea el PCB y el Hilo 0

	var procesoCreado types.PCB
	PIDcount++
	procesoCreado.PID = PIDcount
	procesoCreado.TIDs = []int{0}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	// Se agrega el proceso a NEW
	global.NEW.Add(&procesoCreado)
	// todavia nose donde poner los semaforos
	processToReady(processSize, prioridad)
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
		fmt.Println("Syscall no encontrada: ", syscallName)
	}
}

//ALGUNAS FUNCIONES AUXILIARES
