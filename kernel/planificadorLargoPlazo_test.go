package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"testing"
)

func TestProcessCreate(t *testing.T) {
	// Configurar variables globales para pruebas
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelsync.ChannelProcessArguments = make(chan []string, 1)
	PIDcount = 0 // Reiniciar el contador de PID para la prueba

	// Definir los argumentos para el proceso
	args := []string{"test_file", "500", "1"}

	// Llamar a la syscall ProcessCreate
	err := ProcessCreate(args)
	if err != nil {
		t.Errorf("Error inesperado al crear el proceso: %v", err)
	}

	// Verificar que los argumentos se hayan enviado al canal
	select {
	case receivedArgs := <-kernelsync.ChannelProcessArguments:
		if len(receivedArgs) != 3 || receivedArgs[0] != "test_file" || receivedArgs[1] != "500" || receivedArgs[2] != "1" {
			t.Errorf("Los argumentos recibidos en el canal no coinciden: %v", receivedArgs)
		}
	default:
		t.Errorf("No se recibieron argumentos en ChannelProcessArguments")
	}

	// Verificar que se haya creado un PCB con PID correcto y que esté en NEW
	if len(kernelglobals.EveryPCBInTheKernel) == 0 {
		t.Errorf("No se ha creado ningún PCB en EveryPCBInTheKernel")
	} else {
		pcb := kernelglobals.EveryPCBInTheKernel[0]
		if pcb.PID != 1 || len(pcb.TIDs) != 1 || pcb.TIDs[0] != 0 {
			t.Errorf("PCB creado incorrectamente. PID: %d, TIDs: %v", pcb.PID, pcb.TIDs)
		}
	}
	logCurrentState("Estado Final")
}

func TestProcessExit(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExecStateThread = nil

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	pcb := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0, 1, 2, 3},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)

	// Obtener la referencia correcta del PCB desde EveryPCBInTheKernel
	fatherPCB := &kernelglobals.EveryPCBInTheKernel[len(kernelglobals.EveryPCBInTheKernel)-1]

	// Crear 4 hilos asociados a este PCB, ahora utilizando la referencia correcta del PCB
	mainThread := kerneltypes.TCB{TID: 0, Prioridad: 1, FatherPCB: fatherPCB}
	readyThread := kerneltypes.TCB{TID: 1, Prioridad: 1, FatherPCB: fatherPCB}
	blockedThread := kerneltypes.TCB{TID: 2, Prioridad: 1, FatherPCB: fatherPCB}
	newThread := kerneltypes.TCB{TID: 3, Prioridad: 1, FatherPCB: fatherPCB}

	// Agregar los hilos a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread, readyThread, blockedThread, newThread)

	// Obtener punteros correctos de cada hilo de EveryTCBInTheKernel
	mainThreadPtr := &kernelglobals.EveryTCBInTheKernel[0]
	readyThreadPtr := &kernelglobals.EveryTCBInTheKernel[1]
	blockedThreadPtr := &kernelglobals.EveryTCBInTheKernel[2]
	newThreadPtr := &kernelglobals.EveryTCBInTheKernel[3]

	// Asignar el hilo principal como el hilo ejecutándose
	kernelglobals.ExecStateThread = mainThreadPtr

	// Agregar el hilo de Ready a ReadyStateQueue usando su puntero
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{},
	}
	kernelglobals.ShortTermScheduler.AddToReady(readyThreadPtr)

	// Agregar el hilo de Blocked a BlockedStateQueue usando su puntero
	kernelglobals.BlockedStateQueue.Add(blockedThreadPtr)

	// Agregar el hilo de New a NewStateQueue usando su puntero
	kernelglobals.NewStateQueue.Add(newThreadPtr)

	logCurrentState("Estado Inicial")

	// Llamar a la syscall ProcessExit
	err := ProcessExit([]string{})
	if err != nil {
		t.Errorf("Error inesperado en ProcessExit: %v", err)
	}

	// Verificar que todos los hilos asociados al PCB fueron movidos a la cola ExitStateQueue
	for _, tcb := range []*kerneltypes.TCB{mainThreadPtr, readyThreadPtr, blockedThreadPtr, newThreadPtr} {
		hiloEncontrado := false
		kernelglobals.ExitStateQueue.Do(func(exitTCB *kerneltypes.TCB) {
			if exitTCB.TID == tcb.TID && exitTCB.FatherPCB.PID == pcb.PID {
				hiloEncontrado = true
			}
		})

		if !hiloEncontrado {
			t.Errorf("El TID <%d> del PCB con PID <%d> no fue movido correctamente a ExitStateQueue", tcb.TID, pcb.PID)
		}
	}

	logCurrentState("Estado luego de ProcessExit")
}
