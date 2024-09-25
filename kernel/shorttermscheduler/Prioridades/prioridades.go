package Prioridades

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Prioridades struct {
	readyThreads []kerneltypes.TCB
}

func (prioridades *Prioridades) ThreadExists(tid int, pid int) (bool, error) {
	for _, v := range prioridades.readyThreads {
		if v.TID == tid && v.ConectPCB.PID == pid {
			return true, nil
		}
	}

	return false, errors.New("hilo no encontrado en la cola de prioridades o no pertenece al PCB con PID especificado")
}

func (prioridades *Prioridades) ThreadRemove(tid int, pid int) error {
	existe, err := prioridades.ThreadExists(tid, pid)
	if err != nil {
		return err
	}

	if !existe {
		return errors.New("el hilo con el TID especificado no se encontró en la cola de prioridades o no pertenece al PCB con PID especificado")
	}

	for i, v := range prioridades.readyThreads {
		if v.TID == tid && v.ConectPCB.PID == pid {
			prioridades.readyThreads = append(prioridades.readyThreads[:i], prioridades.readyThreads[i+1:]...)
			logger.Info("Hilo con TID <%d> del PCB con PID <%d> eliminado de la cola de prioridades", tid, pid)
			return nil
		}
	}

	return errors.New("el hilo con el TID especificado no se encontró en la cola de prioridades después de la verificación")
}

func (prioridades *Prioridades) Planificar() (kerneltypes.TCB, error) {
	<-kernelsync.PendingThreadsChannel

	selectedProces := prioridades.readyThreads[0]
	// El proceso se quita de la cola, si por alguna razón el proceso vuelve de CPU sin terminar, debería "creárselo"
	// de nuevo y agregarlo a la cola. TODO: Cómo rompe esto el tema del quantum??
	prioridades.readyThreads = prioridades.readyThreads[1:]

	return selectedProces, nil
}

func (prioridades *Prioridades) AddToReady(threadToAdd kerneltypes.TCB) error {
	logger.Trace("Adding thread to ready (Prioridades): %v", threadToAdd)

	// Si es la primera vez que se llama a la función (la lista es nula), creala
	if prioridades.readyThreads == nil {
		logger.Trace("Creating slice of ready threads")
		prioridades.readyThreads = make([]kerneltypes.TCB, 0)
	}

	inserted := false
	// Por cada hilo que ya está en la lista
	for i := range prioridades.readyThreads {
		// Si la prioridad del hilo a agregar es mayor a lo que acabamos de leer
		if threadToAdd.Prioridad < prioridades.readyThreads[i].Prioridad {
			// Entonces, insertá el hilo en orden
			prioridades.readyThreads = append(prioridades.readyThreads[:i+1], prioridades.readyThreads[i:]...)
			prioridades.readyThreads[i] = threadToAdd
			inserted = true
			break
		}
	}

	if !inserted {
		prioridades.readyThreads = append(prioridades.readyThreads, threadToAdd)
	}

	go func() {
		kernelsync.PendingThreadsChannel <- true
	}()

	// Si es necesario, desalojá la cpu
	if threadToAdd.Prioridad < kernelglobals.ExecStateThread.Prioridad {
		err := shorttermscheduler.CpuInterrupt(
			types.Interruption{
				Type: types.InterruptionEviction,
			})
		if err != nil {
			return err
		}
	}

	logger.Trace("Slice left like this: %v", prioridades.readyThreads)

	return nil
}
