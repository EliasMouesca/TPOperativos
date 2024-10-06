package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
}

// Función auxiliar para registrar el estado actual de ExecStateThread y de los hilos bloqueados
func logCurrentState(context string) {
	logger.Info("\n### %s ###", context)

	// Mostrar estados de PCBs
	logger.Info("## -------- ESTADOS DE TCBs Y PCBs -------- ")
	logger.Info(" - PCBs -")
	for _, pcb := range kernelglobals.EveryPCBInTheKernel {
		logger.Info("PCB PID: %d, TIDs: %v", pcb.PID, pcb.TIDs)
		logger.Info("  Mutexes: ")
		for _, mutex := range pcb.CreatedMutexes {
			assignedTID := types.Tid(-1)
			if mutex.AssignedTCB != nil {
				assignedTID = mutex.AssignedTCB.TID
			}
			logger.Info("    - %s : %d", mutex.Name, assignedTID)
		}
	}

	// Mostrar estados de TCBs
	logger.Info(" - TCBs -")
	for _, tcb := range kernelglobals.EveryTCBInTheKernel {
		logger.Info("(<%d:%d>), Prioridad: %d", tcb.FatherPCB.PID, tcb.TID, tcb.Prioridad)

		if len(tcb.LockedMutexes) == 0 {
			logger.Info("  No hay mutexes bloqueados por este TCB")
		} else {
			logger.Info("  Mutexes locked por TCB:")
			for _, lockedMutex := range tcb.LockedMutexes {
				logger.Info("    - %s", lockedMutex.Name)
			}
		}
	}

	logger.Info("\n")

	// Mostrar estados de las colas
	logger.Info("## -------- ESTADOS DE COLAS -------- ")

	// Mostrar la cola de NewStateQueue
	logger.Info("NewStateQueue: ")
	kernelglobals.NewStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  (<%d:%d>)", tcb.FatherPCB.PID, tcb.TID)
	})

	// Mostrar la cola de BlockedStateQueue
	logger.Info("BlockedStateQueue: ")
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  (<%d:%d>)", tcb.FatherPCB.PID, tcb.TID)
	})

	// Mostrar la cola de ExitStateQueue
	logger.Info("ExitStateQueue: ")
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logger.Info("  (<%d:%d>)", tcb.FatherPCB.PID, tcb.TID)
	})
	logger.Info("\n")

	// Validar que ExecStateThread no sea nil
	if kernelglobals.ExecStateThread != nil {
		logger.Info("Estado actual de ExecStateThread: PID <%d>, TID <%d>, LockedMutexes: ",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
		)
		if len(kernelglobals.ExecStateThread.LockedMutexes) == 0 {
			logger.Info("  No hay mutexes bloqueados por el hilo en ejecución")
		} else {
			for _, mutex := range kernelglobals.ExecStateThread.LockedMutexes {
				logger.Info("	-%v", mutex.Name)
			}
		}
	} else {
		logger.Info("No hay hilo en ejecución actualmente.")
	}
}
