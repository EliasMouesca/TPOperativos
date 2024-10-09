package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"testing"
)

// ---- TODO: PARA LOS TEST DE SUCCES MEMORIA DEBE ESTAR EJECUTANDOSE, PARA LOS DE ERROR NO DEBE ESTAR EJECUTANDO. ----

func TestDumpMemory_Success(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola de ready
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0},
	}

	// Crear el TCB (thread) del proceso en ejecución
	execTCB := kerneltypes.TCB{
		TID:           0,
		Prioridad:     1,
		FatherPCB:     &newPCB,
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[0] // Hilo ejecutándose

	logCurrentState("Estado Inicial")

	// Llamar a DumpMemory con un request válido
	err := DumpMemory([]string{})
	if err != nil {
		t.Errorf("Error inesperado en DumpMemory: %v", err)
	}

	// Verificar que el TCB fue removido de la cola de bloqueados y movido a ready
	existsInBlocked := false
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == execTCB.TID {
			existsInBlocked = true
		}
	})

	if existsInBlocked {
		t.Errorf("El TCB no fue removido de la cola de bloqueados correctamente")
	}

	existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(execTCB.TID, newPID)
	if !existsInReady {
		t.Errorf("El TCB no fue añadido a la cola de ready correctamente")
	}

	logCurrentState("Estado Final")
}

func TestDumpMemory_Success_MultipleThreads(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola de ready
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0, 1, 2, 3}, // Añadimos 4 hilos
	}

	// Agregar el PCB a EveryPCBInTheKernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Crear el TCB (thread) del proceso en ejecución
	execTCB := kerneltypes.TCB{
		TID:           0,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0], // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[0] // Hilo ejecutándose

	// Crear hilos adicionales y agregar a colas correspondientes
	readyTCB := kerneltypes.TCB{
		TID:           1,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}
	blockedTCB := kerneltypes.TCB{
		TID:           2,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}
	newTCB := kerneltypes.TCB{
		TID:           3,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}

	// Agregar los TCBs a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, readyTCB, blockedTCB, newTCB)

	// Agregar los TCBs a sus respectivas colas
	kernelglobals.ShortTermScheduler.AddToReady(&kernelglobals.EveryTCBInTheKernel[1]) // readyTCB
	kernelglobals.BlockedStateQueue.Add(&kernelglobals.EveryTCBInTheKernel[2])         // blockedTCB
	kernelglobals.NewStateQueue.Add(&kernelglobals.EveryTCBInTheKernel[3])             // newTCB

	logCurrentState("Estado Inicial con múltiples hilos")

	// Llamar a DumpMemory con un request válido
	err := DumpMemory([]string{})
	if err != nil {
		t.Errorf("Error inesperado en DumpMemory: %v", err)
	}

	// Verificar que el hilo en ejecución fue removido de BlockedStateQueue y movido a ReadyStateQueue
	existsInBlocked := false
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == execTCB.TID {
			existsInBlocked = true
		}
	})

	if existsInBlocked {
		t.Errorf("El TCB no fue removido de la cola de bloqueados correctamente")
	}

	existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(execTCB.TID, newPID)
	if !existsInReady {
		t.Errorf("El TCB no fue añadido a la cola de ready correctamente")
	}

	// Verificar que los otros hilos permanecen en sus colas correspondientes
	existsInReady, _ = kernelglobals.ShortTermScheduler.ThreadExists(readyTCB.TID, newPID)
	if !existsInReady {
		t.Errorf("El TCB readyTCB no está en la cola de ready")
	}

	existsInBlocked = false
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == blockedTCB.TID {
			existsInBlocked = true
		}
	})
	if !existsInBlocked {
		t.Errorf("El TCB blockedTCB no está en la cola de bloqueados")
	}

	existsInNew := false
	kernelglobals.NewStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == newTCB.TID {
			existsInNew = true
		}
	})
	if !existsInNew {
		t.Errorf("El TCB newTCB no está en la cola de new")
	}

	logCurrentState("Estado Final con múltiples hilos")
}

// ESTE TEST DEBE SER EJECUTADO SIN QUE SE EJECUTE: memoria.go
func TestDumpMemory_Error(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola de ready
	}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{} // Cola de Exit

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0},
	}

	// Agregar el PCB a EveryPCBInTheKernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Crear el TCB (thread) del proceso en ejecución
	execTCB := kerneltypes.TCB{
		TID:           0,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0], // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}

	// Agregar el TCB a EveryTCBInTheKernel y ponerlo como el hilo en ejecución
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[0] // Hilo ejecutándose

	logCurrentState("Estado Inicial")

	// No se inicia el servidor de memoria, por lo que el request fallará
	err := DumpMemory([]string{})
	if err == nil {
		t.Errorf("Se esperaba un error en DumpMemory, pero no ocurrió")
	}

	logCurrentState("Estado despues de DempMemory")

	// Verificar que el TCB fue movido a ExitStateQueue debido al error
	foundInExitQueue := false
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == execTCB.TID {
			foundInExitQueue = true
		}
	})

	if !foundInExitQueue {
		t.Errorf("El TCB no fue movido a la cola de Exit correctamente")
	}

	// Verificar que ExecStateThread sea nil después del error
	if kernelglobals.ExecStateThread != nil {
		t.Errorf("ExecStateThread debería ser nil, pero no lo es")
	}

	logCurrentState("Estado Final")
}

func TestDumpMemory_Error_MultipleThreads(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola de ready
	}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{} // Cola de Exit

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0, 1, 2, 3}, // Añadimos 4 hilos (incluyendo el de ejecución)
	}

	// Agregar el PCB a EveryPCBInTheKernel
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Crear el TCB (thread) del proceso en ejecución
	execTCB := kerneltypes.TCB{
		TID:           0,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0], // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[0] // Hilo ejecutándose

	// Crear hilos adicionales y agregar a colas correspondientes
	readyTCB := kerneltypes.TCB{
		TID:           1,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}
	blockedTCB := kerneltypes.TCB{
		TID:           2,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}
	newTCB := kerneltypes.TCB{
		TID:           3,
		Prioridad:     1,
		FatherPCB:     &kernelglobals.EveryPCBInTheKernel[0],
		LockedMutexes: []*kerneltypes.Mutex{},
	}

	// Agregar los TCBs a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, readyTCB, blockedTCB, newTCB)

	// Agregar los TCBs a sus respectivas colas
	kernelglobals.ShortTermScheduler.AddToReady(&kernelglobals.EveryTCBInTheKernel[1]) // readyTCB
	kernelglobals.BlockedStateQueue.Add(&kernelglobals.EveryTCBInTheKernel[2])         // blockedTCB
	kernelglobals.NewStateQueue.Add(&kernelglobals.EveryTCBInTheKernel[3])             // newTCB

	logCurrentState("Estado Inicial con múltiples hilos")

	// No se inicia el servidor de memoria, por lo que el request fallará
	err := DumpMemory([]string{})
	if err == nil {
		t.Errorf("Se esperaba un error en DumpMemory, pero no ocurrió")
	}

	logCurrentState("Estado despues de DumpMemory con error")

	// Verificar que todos los TCBs fueron movidos a ExitStateQueue
	for _, tcb := range []*kerneltypes.TCB{&execTCB, &readyTCB, &blockedTCB, &newTCB} {
		foundInExitQueue := false
		kernelglobals.ExitStateQueue.Do(func(exitTCB *kerneltypes.TCB) {
			if exitTCB.TID == tcb.TID {
				foundInExitQueue = true
			}
		})

		if !foundInExitQueue {
			t.Errorf("El TCB con TID <%d> no fue movido a la cola de Exit correctamente", tcb.TID)
		}
	}

	// Verificar que ExecStateThread sea nil después del error
	if kernelglobals.ExecStateThread != nil {
		t.Errorf("ExecStateThread debería ser nil, pero no lo es")
	}

	logCurrentState("Estado Final con múltiples hilos")
}
