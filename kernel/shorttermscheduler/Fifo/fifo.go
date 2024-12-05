package Fifo

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Fifo struct {
	Ready         types.Queue[*kerneltypes.TCB]
	LastScheduled *kerneltypes.TCB
}

func (f *Fifo) Init() {
}

func (f *Fifo) ThreadExists(tid types.Tid, pid types.Pid) (bool, error) {
	for _, v := range f.Ready.GetElements() {
		if v.TID == tid && v.FatherPCB.PID == pid {
			return true, nil
		}
	}
	return false, errors.New("hilo no encontrado o no pertenece al PCB con el PID especificado")
}

func (f *Fifo) ThreadRemove(tid types.Tid, pid types.Pid) error {
	existe, err := f.ThreadExists(tid, pid)
	if err != nil {
		return err
	}

	if existe {
		queueSize := f.Ready.Size()
		for i := 0; i < queueSize; i++ {
			v, err := f.Ready.GetAndRemoveNext()
			if err != nil {
				return err
			}

			// Volver a agregar el TCB solo si no coincide con el tid y pid
			if v.TID != tid || v.FatherPCB.PID != pid {
				f.Ready.Add(v)
			} else {
				go func() {
					<-kernelsync.PendingThreadsChannel
				}()
			}

		}
		return nil
	}

	return errors.New("hilo no encontrado o no pertenece al PCB con el PID especificado")
}

// Planificar devuelve el próximo hilo a ejecutar o error en función del algoritmo FIFO
// es una función que se bloquea si no hay procesos listos y se desbloquea sola si llegan a venir nuevos procesos listos
func (f *Fifo) Planificar() (*kerneltypes.TCB, error) {
	var nextTcb *kerneltypes.TCB
	var err error

	logger.Trace("%s", f.Ready)
	// Fifo lo único que hace para seleccionar procesos es tomar el primero que entró
	nextTcb, err = f.Ready.GetAndRemoveNext()
	if err != nil {
		return nil, errors.New("se quiso obtener un hilo y no habia ningun hilo en ready")
	}

	f.LastScheduled = nextTcb

	logger.Debug("FIFO Elijió el hilo (<%v>:<%v>)", nextTcb.FatherPCB.PID, nextTcb.TID)
	// Retorná el hilo elegido
	return nextTcb, nil
}

// AddToReady Le avisa al STS (versión FIFO) que hay un nuevo proceso listo
func (f *Fifo) AddToReady(tcb *kerneltypes.TCB) error {
	// Agregá el proceso a la cola fifo
	f.Ready.Add(tcb)

	// Mandá mensaje por el canal, o sea, permití que una vuelta más de Planificar() ejecute
	go func() {
		logger.Trace("Hay hilos en ready en FIFO")
		logger.Trace("%s", f.Ready)
		kernelsync.PendingThreadsChannel <- true
	}()

	return nil
}
