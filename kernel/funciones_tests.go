package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
}

// Función auxiliar para registrar el estado actual de ExecStateThread y de los hilos bloqueados
func logCurrentState(context string) {
	logger.Info("\n### %s ###", context)

	logger.Info("## -------- ESTADOS DE TCBs Y PCBs -------- ")
	logger.Info("Estado actual de EveryPCBInTheKernel: %v", kernelglobals.EveryPCBInTheKernel)
	logger.Info("Estado actual de EveryTCBInTheKernel: %v \n", kernelglobals.EveryTCBInTheKernel)

	logger.Info("## -------- ESTADOS DE COLAS -------- ")
	logger.Info("Estado actual de BlockedStateQueue: %v", kernelglobals.BlockedStateQueue)
	logger.Info("Estado actual de ExitStateQueue: %v \n", kernelglobals.ExitStateQueue)

	// Validar que ExecStateThread no sea nil
	if kernelglobals.ExecStateThread != nil {
		logger.Info("Estado actual de ExecStateThread: PID <%d>, TID <%d>, CreatedMutexes: %v",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
			kernelglobals.ExecStateThread.LockedMutexes,
		)
	} else {
		logger.Info("No hay hilo en ejecución actualmente.")
	}
	/*
		// Verificar existencia de hilos en la cola de ready
		for _, tcb := range kernelglobals.EveryTCBInTheKernel {
			exists, err := kernelglobals.ShortTermScheduler.ThreadExists(tcb.TID, tcb.FatherPCB.PID)
			if err != nil {
				logger.Error("Error al verificar existencia del TCB en la cola de Ready: %v", err)
				continue
			}
			if exists {
				logger.Info("  TID <%d> del PCB con PID <%d>, Prioridad <%d>, Mutexes: %v",
					tcb.TID, tcb.FatherPCB.PID, tcb.Prioridad, tcb.LockedMutexes)
			}
		}

		// Estado de BlockedStateQueue
		logger.Info("Estado actual de BlockedStateQueue:")
		kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
			logger.Info("  TID <%d> del PCB con PID <%d>, Prioridad <%d>, Mutexes: %v",
				tcb.TID, tcb.FatherPCB.PID, tcb.Prioridad, tcb.LockedMutexes)
		})

		// Estado de ExitStateQueue
		logger.Info("Estado actual de ExitStateQueue:")
		kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
			logger.Info("  TID <%d> del PCB con PID <%d>",
				tcb.TID, tcb.FatherPCB.PID)
		})

		// Información sobre los mutexes en GlobalMutexRegistry
		for mutexID, mutexWrapper := range kernelglobals.ExecStateThread.FatherPCB.CreatedMutexes {
			logger.Info("Estado del Mutex ID <%v>: AssignedTID <%v>, BlockedTCBs: [", mutexID, mutexWrapper.AssignedTCB)
			for _, blockedTCB := range mutexWrapper.BlockedTCBs {
				logger.Info("  TID <%d> del PCB con PID <%d>", blockedTCB.TID, blockedTCB.FatherPCB.PID)
			}
			logger.Info("]")
		}*/
}

/*
// Función auxiliar para registrar el estado actual de un PCB
func logPCBState(context string, pcb *kerneltypes.PCB) {
	logger.Info("### %s ###", context)
	logger.Info("Estado actual del PCB: PID <%d>", pcb.PID)

	// Mostrar información de los TIDs asociados al PCB
	logger.Info("	Lista de TIDs asociados al PCB con PID <%d>: %v", pcb.PID, pcb.TIDs)

	// Recorrer los TCBs en la ReadyStateQueue para obtener información detallada de los hilos
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.FatherPCB == pcb {
			logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v", tcb.TID, tcb.Prioridad, tcb.LockedMutexes)
		}
	})
	if kernelglobals.ExecStateThread.FatherPCB == pcb {
		logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v",
			kernelglobals.ExecStateThread.TID, kernelglobals.ExecStateThread.Prioridad, kernelglobals.ExecStateThread.LockedMutexes)
	}
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.FatherPCB == pcb {
			logger.Info("		TCB -> TID <%d>: Prioridad <%d>, Mutexes: %v", tcb.TID, tcb.Prioridad, tcb.LockedMutexes)
		}
	})
	// Mostrar información de los mutexes asociados al PCB
	logger.Info("	Mutexes asociados al PCB con PID <%d>: %v", pcb.PID, pcb.CreatedMutexes)
}
*/
