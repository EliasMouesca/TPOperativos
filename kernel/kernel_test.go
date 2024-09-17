package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
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
	args := []string{"testfile", "1024", "1"}
	syscall := syscalls.Syscall{
		Type:      2,
		Arguments: args,
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		ExecuteSyscall(syscall)
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToReady()
	}()

	// Esperamos a que finalicen todas las rutinas
	kernelsync.WaitPlanificadorLP.Wait()
}
