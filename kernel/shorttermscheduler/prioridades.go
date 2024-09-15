package shorttermscheduler

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Prioridades struct {
	readyThreads []kerneltypes.TCB
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
		// If the thread wasn't inserted in the loop, append it at the end
		prioridades.readyThreads = append(prioridades.readyThreads, threadToAdd)
	}

	go func() {
		kernelsync.PendingThreadsChannel <- true
	}()

	logger.Trace("Slice left like this: %v", prioridades.readyThreads)

	return nil
}
