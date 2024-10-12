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
	"os"
)

// ANSI escape codes for colors
const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Bold    = "\033[1m"
)

func logCurrentState(context string) {
	if false {
		// Crear un buffer para acumular todo el log
		var logBuffer bytes.Buffer

		logBuffer.WriteString(fmt.Sprintf("\n### %s ###\n", context))

		// Mostrar estados de PCBs
		logBuffer.WriteString(fmt.Sprintf("%s## -------- ESTADOS DE TCBs Y PCBs -------- %s\n", Cyan, Reset))
		logBuffer.WriteString(fmt.Sprintf("%s - PCBs -%s\n", Yellow, Reset))
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
					if assignedTID == -1 {
						logBuffer.WriteString(fmt.Sprintf("	- %s : nil\n", mutex.Name))
					} else {
						logBuffer.WriteString(fmt.Sprintf("	- %s : (<%v:%d>)\n", mutex.Name, pcb.PID, assignedTID))
					}
				}
			}
		}

		// Mostrar estados de TCBs
		logBuffer.WriteString(fmt.Sprintf("%s - TCBs -%s\n", Yellow, Reset))
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
		logBuffer.WriteString(fmt.Sprintf("%s## -------- ESTADOS DE COLAS Y TCB EJECUTANDO -------- %s\n", Cyan, Reset))

		// Mostrar la cola de NewStateQueue
		logBuffer.WriteString(fmt.Sprintf("%sNewStateQueue:%s\n", Green, Reset))
		kernelglobals.NewStateQueue.Do(func(tcb *kerneltypes.TCB) {
			logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
		})

		// Mostrar la cola de Ready según el planificador
		switch scheduler := kernelglobals.ShortTermScheduler.(type) {
		case *Fifo.Fifo:
			logBuffer.WriteString(fmt.Sprintf("%sReadyStateQueue FIFO:%s\n", Green, Reset))
			scheduler.Ready.Do(func(tcb *kerneltypes.TCB) {
				logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
			})
		case *Prioridades.Prioridades:
			logBuffer.WriteString(fmt.Sprintf("%sReadyStateQueue PRIORIDADES:%s\n", Green, Reset))
			for _, tcb := range scheduler.ReadyThreads {
				logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
			}
		case *ColasMultinivel.ColasMultiNivel:
			logBuffer.WriteString(fmt.Sprintf("%sReadyStateQueue MULTI NIVEL:%s\n", Green, Reset))
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
		logBuffer.WriteString(fmt.Sprintf("%sBlockedStateQueue:%s\n", Green, Reset))
		kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
			logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
		})

		// Mostrar la cola de ExitStateQueue
		logBuffer.WriteString(fmt.Sprintf("%sExitStateQueue:%s\n", Green, Reset))
		kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
			logBuffer.WriteString(fmt.Sprintf("  (<%d:%d>)\n", tcb.FatherPCB.PID, tcb.TID))
		})

		// Mostrar el hilo en ejecución
		if kernelglobals.ExecStateThread != nil {
			logBuffer.WriteString(fmt.Sprintf("%sExecStateThread:%s\n", Green, Reset))
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
		fmt.Fprint(os.Stdout, logBuffer.String())

	}
}
