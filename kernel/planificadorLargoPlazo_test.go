package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
	"time"
)

func TestNewProcessToReady(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewPCBStateQueue = types.Queue[*kerneltypes.PCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola Ready
	}

	// Configurar la dirección y puerto de la memoria real
	kernelglobals.Config.MemoryAddress = "localhost" // O IP real
	kernelglobals.Config.MemoryPort = 8080           // El puerto real de memoria

	// Crear el PCB y agregarlo a la cola NewPCBStateQueue
	PIDcount++
	processCreate := kerneltypes.PCB{
		PID:  types.Pid(PIDcount),
		TIDs: []types.Tid{0},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, processCreate)
	fatherPCB := &kernelglobals.EveryPCBInTheKernel[len(kernelglobals.EveryPCBInTheKernel)-1]

	// Crear los argumentos que se enviarán a través del canal
	args := []string{"test_file", "500", "1"}

	// Enviar los argumentos a través del canal
	kernelsync.ChannelProcessArguments = make(chan []string, 1)
	kernelsync.ChannelProcessArguments <- args

	// Añadir el PCB a NewPCBStateQueue
	kernelglobals.NewPCBStateQueue.Add(fatherPCB)

	logCurrentState("Antes de pasar el proceso a READY")

	// Llamar a la función en un goroutine para simular el comportamiento concurrente
	go NewProcessToReady()

	// Esperar a que el proceso sea movido a la cola Ready
	time.Sleep(100 * time.Millisecond) // Simulación de latencia

	// Verificar que el hilo principal fue movido a la cola Ready
	existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(0, processCreate.PID)
	if !existsInReady {
		t.Errorf("El TID <0> del PCB con PID <%d> no fue movido correctamente a ReadyStateQueue", processCreate.PID)
	} else {
		logger.Info("## (<%v:%v>) fue movido a la cola de READY.", processCreate.PID, processCreate.TIDs[0])
	}

	// Verificar que el TCB fue agregado a EveryTCBInTheKernel
	if len(kernelglobals.EveryTCBInTheKernel) != 1 {
		t.Errorf("No se agregó correctamente el TCB a EveryTCBInTheKernel")
	}

	// Verificar que el PCB ya no está en la cola NewPCBStateQueue
	if !kernelglobals.NewPCBStateQueue.IsEmpty() {
		t.Errorf("El PCB con PID <%d> aún se encuentra en la cola NewPCBStateQueue", processCreate.PID)
	}

	logCurrentState("Estado Final luego de mover el proceso a Ready")
}
