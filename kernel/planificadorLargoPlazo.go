package main

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

var available int = 0
var NEW []types.PCB
var Ready []types.TCB

func planificadorLargoPlazo(syscall syscalls.Syscall) {
	syscallName := syscalls.SyscallNames[syscall.Type]
	logger.Info("\n\n## (<PID>:<TID>) - Solicitó syscall: <%v>", syscallName)

	switch syscall.Type {

	//PROCESS_CREATE
	case syscalls.ProcessCreate:
		pseudocodigo := syscall.Arguments[0]
		processSize, _ := strconv.Atoi(syscall.Arguments[1])
		prioridad, _ := strconv.Atoi(syscall.Arguments[2])

		if len(NEW) != 0 {
			NEW = append(NEW, types.PCB{})
			// Si NEW no esta vacio significa que hay un proceso esperando a ser mandado a Ready
			// Habria que hacer una sincronizacion de los procesos que vayan llegando
			// y que vayan preguntando a memoria a medida que pasan los procesos a Ready
		}

		PROCESS_CREATE(pseudocodigo, processSize, prioridad)

	//CREATE_THREAD
	case syscalls.ThreadCreate:
		pid, _ := strconv.Atoi(syscall.Arguments[0])
		pseudocodigo := syscall.Arguments[1]
		prioridad, _ := strconv.Atoi(syscall.Arguments[2])

		// Encontrar el PCB correspondiente
		var proceso *types.PCB
		for i := range NEW {
			if NEW[i].PID == pid {
				proceso = &NEW[i]
				break
			}
		}

		if proceso != nil {
			THREAD_CREATE(proceso, pseudocodigo, prioridad)
			logger.Info("## (%d:<TID>) Se crea el hilo - Estado: READY", proceso.PID)
		} else {
			logger.Error("No se encontró el proceso con PID <%d> en la lista NEW", pid)
		}

	//THREAD_EXIT
	case syscalls.ThreadExit:

	//PROCESS_EXIT
	case syscalls.ProcessExit:
		// Suponiendo que el PID se pasa como primer argumento
		pid, _ := strconv.Atoi(syscall.Arguments[0])

		// Encontrar el PCB correspondiente
		var procesoAEliminar *types.PCB
		for i := range NEW {
			if NEW[i].PID == pid {
				procesoAEliminar = &NEW[i]
				break
			}
		}
		if procesoAEliminar != nil {
			PROCESS_EXIT(*procesoAEliminar)
		} else {
			logger.Error("No se encontró el proceso con PID <%d> en la lista NEW", pid)
		}

	}
}
