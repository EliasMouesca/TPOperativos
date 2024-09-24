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

func (prioridades *Prioridades) ThreadExists(thread types.Thread) (bool, error) {
	for _, v := range prioridades.readyThreads {
		if v.TID == thread.Tid {
			return true, nil
		}
	}
	return false, errors.New("hilo no encontrado")
}

func (prioridades *Prioridades) ThreadRemove(thread types.Thread) error {
	existe, err := prioridades.ThreadExists(thread)
	if err != nil {
		return err
	}

	if existe {
		for i, v := range prioridades.readyThreads {
			if v.TID != thread.Tid {
				copy(prioridades.readyThreads[i:], prioridades.readyThreads[i+1:])
				prioridades.readyThreads = prioridades.readyThreads[:len(prioridades.readyThreads)-1]
				return nil
			}
		}
	} else {
		return errors.New("el hilo pedido no existe :/")
	}

	return nil

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
