package Fifo

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
)

type Fifo struct {
	ready types.Queue[kerneltypes.TCB]
}

func (f *Fifo) ThreadExists(thread types.Thread) (bool, error) {
	for _, v := range f.ready.GetElements() {
		if v.TID == thread.Tid {
			return true, nil
		}
	}
	return false, errors.New("hilo no encontrado")
}

func (f *Fifo) ThreadRemove(thread types.Thread) error {
	existe, err := f.ThreadExists(thread)
	if err != nil {
		return err
	}

	// TODO: Cursed
	if existe {
		for !f.ready.IsEmpty() {
			v, err := f.ready.GetAndRemoveNext()
			if err != nil {
				return err
			}

			if v.TID != thread.Tid {
				f.ready.Add(v)
			}
		}
	} else {
		return errors.New("el hilo pedido no existe :/")
	}

	return nil

}

// Planificar devuelve el próximo hilo a ejecutar o error en función del algoritmo FIFO
// es una función que se bloquea si no hay procesos listos y se desbloquea sola si llegan a venir nuevos procesos listos
func (f *Fifo) Planificar() (kerneltypes.TCB, error) {
	// Bloqueate hasta que alguien te mande algo por este channel -> quién manda por este channel? -> AddToReady()
	// Entonces, bloqueate hasta que alguien agregue un hilo a ready.
	<-kernelsync.PendingThreadsChannel

	var nextTcb *kerneltypes.TCB
	var err error

	// Fifo lo único que hace para seleccionar procesos es tomar el primero que entró
	nextTcb, err = f.ready.GetAndRemoveNext()
	if err != nil {
		return kerneltypes.TCB{}, errors.New("se quiso obtener un hilo y no habia ningun hilo en ready")
	}

	// Retorná el hilo elegido
	return *nextTcb, nil
}

// AddToReady Le avisa al STS (versión FIFO) que hay un nuevo proceso listo
func (f *Fifo) AddToReady(tcb kerneltypes.TCB) error {
	// Agregá el proceso a la cola fifo
	f.ready.Add(&tcb)

	// Mandá mensaje por el canal, o sea, permití que una vuelta más de Planificar() ejecute
	go func() {
		kernelsync.PendingThreadsChannel <- true
	}()
	return nil
}
