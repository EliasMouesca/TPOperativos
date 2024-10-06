package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
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
