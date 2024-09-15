package main

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type syscallFunc func(args ...interface{}) error

// TODO: Dónde ponemos esto? en qué carpeta?

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

func ExecuteSyscall(syscallName string, args ...interface{}) error {
	if syscallFunc, exists := syscallSet[syscallName]; exists {
		err := syscallFunc(args...)
		return err
	} else {
		logger.Error("Syscall no encontrada: %v", syscallName)
	}

	return nil
}

var PIDcount int = 0

func PROCESS_CREATE(args ...interface{}) error {
	// Agregar New.Errors
	//pseudoCodigo := args[0]
	//processSize := args[1].(int)
	prioridad := args[2].(int)

	// Se crea el PCB y el Hilo 0

	var procesoCreado kerneltypes.PCB
	PIDcount++
	procesoCreado.PID = PIDcount
	_ = kerneltypes.TCB{
		TID:       0,
		Prioridad: prioridad,
	}
	//procesoCreado.TIDs = []TCB{hiloMain}

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	// TODO: El proceso no debería agregarse a new, debería ser el hilo ? -eli
	// Se agrega el proceso a NEW
	kernelglobals.NewStateQueue.Add(&procesoCreado)

	return nil
}

func PROCESS_EXIT(args ...interface{}) error {
	// nose si estara bien pero el valor TCB ya esta en el canal
	tcb := <-kernelglobals.EXIT // sino usar una lista de un elemento consultar con los pibes
	if tcb.TID == 0 {           // tiene que ser el hiloMain
		conectedProcess := tcb.ConectPCB
		processToExit(conectedProcess)
	} else {
		return errors.New("El hilo que quizo eliminar el proceso, no es el hilo main")
	}

	return nil
}

func THREAD_CREATE(args ...interface{}) error {
	// len(Ready) forma autoincremental ajustable
	// TIDcount forma autoincremental crecientei ndeterminadamente
	// nose que opcion es mejor
	fmt.Println("Creando hilo...")

	return nil
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
