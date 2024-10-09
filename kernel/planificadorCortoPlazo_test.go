package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
	"testing"
)

func TestPlanificadorCortoPlazo(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	tids := make([]types.Tid, 3)
	tids[0] = 1
	tids[1] = 2
	tids[2] = 3
	pcb := kerneltypes.PCB{
		PID:            1,
		TIDs:           tids,
		CreatedMutexes: make([]kerneltypes.Mutex, 0),
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)
	tcb := kerneltypes.TCB{
		Prioridad:     1,
		TID:           1,
		FatherPCB:     &pcb,
		LockedMutexes: make([]*kerneltypes.Mutex, 0),
	}
	tcb1 := kerneltypes.TCB{
		Prioridad:     0,
		TID:           2,
		FatherPCB:     &pcb,
		LockedMutexes: make([]*kerneltypes.Mutex, 0),
	}
	tcb2 := kerneltypes.TCB{
		Prioridad:     2,
		TID:           3,
		FatherPCB:     &pcb,
		LockedMutexes: make([]*kerneltypes.Mutex, 0),
	}

	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, tcb, tcb1, tcb2)
	logCurrentState("Estado antes de inicar corto plazo")
	go func() {
		defer wg.Done()
		planificadorCortoPlazo()
	}()
	go func() {
		defer wg.Done()
		logCurrentState("Estado luego de iniciar corto plazo")
		kernelglobals.ShortTermScheduler.AddToReady(&tcb)
		logCurrentState("Estado luego de mandar tcb a ready")
		kernelglobals.ShortTermScheduler.AddToReady(&tcb1)
		kernelglobals.ShortTermScheduler.AddToReady(&tcb2)
		logCurrentState("Estado luego de mandar todos los tcb a ready")
	}()
	wg.Wait()
}
