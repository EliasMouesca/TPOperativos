package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
}

// Función auxiliar para registrar el estado actual de ExecStateThread y de los hilos bloqueados
func logCurrentState(context string) {
	logger.Info("### %s ###", context)
	logger.Info("Estado actual de ExecStateThread: PID <%d>, TID <%d>, CreatedMutexes: %v",
		kernelglobals.ExecStateThread.FatherPCB.PID,
		kernelglobals.ExecStateThread.TID,
		kernelglobals.ExecStateThread.Mutex,
	)
	// Estado de ReadyStateQueue
	logger.Info("Estado actual de ReadyStateQueue:")
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  TID <%d> del PCB con PID <%d>, Prioridad <%d>, Mutexes: %v",
			tcb.TID, tcb.FatherPCB.PID, tcb.Prioridad, tcb.Mutex)
	})

	// Estado de BlockedStateQueue
	logger.Info("Estado actual de BlockedStateQueue:")
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  TID <%d> del PCB con PID <%d>, Prioridad <%d>, Mutexes: %v",
			tcb.TID, tcb.FatherPCB.PID, tcb.Prioridad, tcb.Mutex)
	})

	// Estado de ExitStateQueue
	logger.Info("Estado actual de ExitStateQueue:")
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  TID <%d> del PCB con PID <%d>",
			tcb.TID, tcb.FatherPCB.PID)
	})

	for mutexID, mutexWrapper := range kernelglobals.GlobalMutexRegistry {
		logger.Info("Estado del CreatedMutexes ID <%d>: AssignedTID <%d>, BlockedTCBs: [", mutexID, mutexWrapper.AssignedTID)
		for _, blockedTCB := range mutexWrapper.BlockedTCBs {
			logger.Info("  TID <%d> del PCB con PID <%d>", blockedTCB.TID, blockedTCB.FatherPCB.PID)
		}
		logger.Info("]")
	}
}

// Función auxiliar para registrar el estado actual de un PCB
func logPCBState(context string, pcb *kerneltypes.PCB) {
	logger.Info("### %s ###", context)
	logger.Info("Estado actual del PCB: PID <%d>", pcb.PID)

	// Mostrar información de los TIDs asociados al PCB
	logger.Info("	Lista de TIDs asociados al PCB con PID <%d>: %v", pcb.PID, pcb.TIDs)

	// Recorrer los TCBs en la ReadyStateQueue para obtener información detallada de los hilos
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.FatherPCB == pcb {
			logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v", tcb.TID, tcb.Prioridad, tcb.Mutex)
		}
	})
	if kernelglobals.ExecStateThread.FatherPCB == pcb {
		logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v",
			kernelglobals.ExecStateThread.TID, kernelglobals.ExecStateThread.Prioridad, kernelglobals.ExecStateThread.Mutex)
	}
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.FatherPCB == pcb {
			logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v", tcb.TID, tcb.Prioridad, tcb.Mutex)
		}
	})
	// Mostrar información de los mutexes asociados al PCB
	logger.Info("	Mutexes asociados al PCB con PID <%d>: %v", pcb.PID, pcb.CreatedMutexes)
}
