package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

type SyscallRequest struct {
	Type        int      `json:"type"`
	Arguments   []string `json:"arguments"`
	Description string   `json:"description"`
}

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
}

func TestProcessToReady(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ExecStateThread = tcb

	args := []string{"testfile", "1024", "1"}
	syscall := syscalls.Syscall{
		Type:      2,
		Arguments: args,
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err := ExecuteSyscall(syscall)
		if err != nil {
			logger.Error("%v", err)
		}
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToReady()
	}()
	kernelsync.SemCreateprocess <- 0
	// no se va a conectar con memoria pero ya le
	// estoy dando el visto para que se conecte

	// Esperamos a que finalicen todas las rutinas
	kernelsync.WaitPlanificadorLP.Wait()
}

func TestProcessToExit(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb1 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ReadyStateQueue.Add(&tcb1)
	kernelglobals.ExecStateThread = tcb
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      10,
		Arguments: args,
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err := ExecuteSyscall(syscall)
		if err != nil {
			logger.Error("%v", err)
		}
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToExit()
	}()
	kernelsync.SemFinishprocess <- 0

	// Esperamos a que finalicen todas las rutinas
	kernelsync.WaitPlanificadorLP.Wait()
}

func TestProcessExit(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb1 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ExecStateThread = tcb
	kernelglobals.ReadyStateQueue.Add(&tcb1)
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      10,
		Arguments: args,
	}

	err := ExecuteSyscall(syscall)
	if err != nil {
		logger.Error("%v", err)
	}
}
