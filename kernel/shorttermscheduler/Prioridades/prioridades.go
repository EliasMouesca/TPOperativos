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
	ReadyThreads []*kerneltypes.TCB
}

func (prioridades *Prioridades) ThreadExists(tid types.Tid, pid types.Pid) (bool, error) {
	for _, v := range prioridades.ReadyThreads {
		if v.TID == tid && v.FatherPCB.PID == pid {
			return true, nil
		}
	}

	return false, errors.New("hilo no encontrado en la cola de prioridades o no pertenece al PCB con PID especificado")
}

func (prioridades *Prioridades) ThreadRemove(tid types.Tid, pid types.Pid) error {
	existe, err := prioridades.ThreadExists(tid, pid)
	if err != nil {
		return err
	}

	if !existe {
		return errors.New("el hilo con el TID especificado no se encontró en la cola de prioridades o no pertenece al PCB con PID especificado")
	}

	for i, v := range prioridades.ReadyThreads {
		if v.TID == tid && v.FatherPCB.PID == pid {
			prioridades.ReadyThreads = append(prioridades.ReadyThreads[:i], prioridades.ReadyThreads[i+1:]...)
			kernelglobals.ExitStateQueue.Add(v)
			logger.Info("Hilo con TID <%d> del PCB con PID <%d> eliminado de la cola de prioridades", tid, pid)
			return nil
		}
	}

	return errors.New("el hilo con el TID especificado no se encontró en la cola de prioridades después de la verificación")
}

func (prioridades *Prioridades) Planificar() (*kerneltypes.TCB, error) {

	selectedProces := prioridades.ReadyThreads[0]
	// El proceso se quita de la cola, si por alguna razón el proceso vuelve de CPU sin terminar, debería "creárselo"
	// de nuevo y agregarlo a la cola. TODO: Cómo rompe esto el tema del quantum??
	prioridades.ReadyThreads = prioridades.ReadyThreads[1:]
	logger.Info("Planificando en Prioridades el hilo con TID: %v", selectedProces.TID)
	return selectedProces, nil
}

func (prioridades *Prioridades) AddToReady(threadToAdd *kerneltypes.TCB) error {
	logger.Info("Adding thread to ready (Prioridades) TID: %v, Prioridad: %v", threadToAdd.TID, threadToAdd.Prioridad)

	// Si es la primera vez que se llama a la función (la lista es nula), creala
	if prioridades.ReadyThreads == nil {
		logger.Trace("Creating slice of ready threads")
		prioridades.ReadyThreads = make([]*kerneltypes.TCB, 0)
	}

	inserted := false
	// Por cada hilo que ya está en la lista
	for i := range prioridades.ReadyThreads {
		// Si la prioridad del hilo a agregar es mayor a lo que acabamos de leer
		if threadToAdd.Prioridad < prioridades.ReadyThreads[i].Prioridad {
			// Entonces, insertá el hilo en orden
			prioridades.ReadyThreads = append(prioridades.ReadyThreads[:i+1], prioridades.ReadyThreads[i:]...)
			prioridades.ReadyThreads[i] = threadToAdd
			inserted = true
			break
		}
	}

	if !inserted {
		prioridades.ReadyThreads = append(prioridades.ReadyThreads, threadToAdd)
	}

	go func() {
		kernelsync.PendingThreadsChannel <- true
	}()

	if kernelglobals.ExecStateThread != nil {
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
	}
	logger.Trace("Slice left like this: %v", prioridades.ReadyThreads)

	return nil
}
