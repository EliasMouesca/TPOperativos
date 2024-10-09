package main

import (
	"bytes"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/ColasMultinivel"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Prioridades"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

// Función auxiliar para registrar el estado actual de ExecStateThread y de los hilos bloqueados
func logCurrentState(context string) {
	// Crear un buffer para acumular todo el log
	var logBuffer bytes.Buffer

	logBuffer.WriteString(fmt.Sprintf("\n### %s ###\n", context))

	// Mostrar estados de PCBs
	logBuffer.WriteString("## -------- ESTADOS DE TCBs Y PCBs -------- \n")
	logBuffer.WriteString(" - PCBs -\n")
	for _, pcb := range kernelglobals.EveryPCBInTheKernel {
		logBuffer.WriteString(fmt.Sprintf("	(<%d:%v>), Mutexes: \n", pcb.PID, pcb.TIDs))
		if len(pcb.CreatedMutexes) == 0 {
			logBuffer.WriteString("  	No hay mutexes creados por este PCB\n")
		} else {
			for _, mutex := range pcb.CreatedMutexes {
				assignedTID := types.Tid(-1)
				if mutex.AssignedTCB != nil {
					assignedTID = mutex.AssignedTCB.TID
				}
				logBuffer.WriteString(fmt.Sprintf("	- %s : %d\n", mutex.Name, assignedTID))
			}
		}
	}

	// Mostrar estados de TCBs
	logBuffer.WriteString(" - TCBs -\n")
	for _, tcb := range kernelglobals.EveryTCBInTheKernel {
		if tcb.JoinedTCB == nil {
			logBuffer.WriteString(fmt.Sprintf("    (<%d:%d>), Prioridad: %d, JoinedTCB: nil\n", tcb.FatherPCB.PID, tcb.TID, tcb.Prioridad))
		} else {
			logBuffer.WriteString(fmt.Sprintf("    (<%d:%d>), Prioridad: %d, JoinedTCB: %v\n", tcb.FatherPCB.PID, tcb.TID, tcb.Prioridad, tcb.JoinedTCB.TID))
		}

		if len(tcb.LockedMutexes) == 0 {
			logBuffer.WriteString("  	No hay mutexes bloqueados por este TCB\n")
		} else {
			logBuffer.WriteString("  	Mutexes locked por TCB:\n")
			for _, lockedMutex := range tcb.LockedMutexes {
				logBuffer.WriteString(fmt.Sprintf("    	- %s\n", lockedMutex.Name))
			}
		}
	}

	logBuffer.WriteString("\n")

	// Mostrar estados de las colas
	logBuffer.WriteString("## -------- ESTADOS DE COLAS Y TCB EJECUTANDO -------- \n")

	// Mostrar la cola de NewStateQueue
	logBuffer.WriteString("NewStateQueue: \n")
	kernelglobals.NewStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
	})

	// Mostrar la cola de Ready según el planificador
	switch scheduler := kernelglobals.ShortTermScheduler.(type) {
	case *Fifo.Fifo:
		logBuffer.WriteString("ReadyStateQueue FIFO: \n")
		scheduler.Ready.Do(func(tcb *kerneltypes.TCB) {
			logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
		})
	case *Prioridades.Prioridades:
		logBuffer.WriteString("ReadyStateQueue PRIORIDADES: \n")
		for _, tcb := range scheduler.ReadyThreads {
			logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
		}
	case *ColasMultinivel.ColasMultiNivel:
		logBuffer.WriteString("ReadyStateQueue MULTI NIVEL: \n")
		for i, queue := range scheduler.ReadyQueue {
			logBuffer.WriteString(fmt.Sprintf("Nivel %d:\n", i))
			queue.Do(func(tcb *kerneltypes.TCB) {
				logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
			})
		}
	default:
		logBuffer.WriteString("No se reconoce el tipo de planificador en uso.\n")
	}

	// Mostrar la cola de BlockedStateQueue
	logBuffer.WriteString("BlockedStateQueue: \n")
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
	})

	// Mostrar la cola de ExitStateQueue
	logBuffer.WriteString("ExitStateQueue: \n")
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
	})

	// Mostrar el hilo en ejecución
	if kernelglobals.ExecStateThread != nil {
		logBuffer.WriteString("ExecStateThread:\n")
		logBuffer.WriteString(fmt.Sprintf("	(<%d:%d>), LockedMutexes: \n",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
		))
		if len(kernelglobals.ExecStateThread.LockedMutexes) == 0 {
			logBuffer.WriteString("  	No hay mutexes bloqueados por el hilo en ejecución\n")
		} else {
			for _, mutex := range kernelglobals.ExecStateThread.LockedMutexes {
				logBuffer.WriteString(fmt.Sprintf("	-%v\n", mutex.Name))
			}
		}
	} else {
		logBuffer.WriteString("No hay hilo en ejecución actualmente.\n")
	}

	logBuffer.WriteString("\n")

	// Escribir el log completo en una sola operación
	logger.Info(logBuffer.String())
}
