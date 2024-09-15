package main

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/types"
)

type Fifo struct {
	ready types.Queue[TCB]
}

// Planificar devuelve el próximo hilo a ejecutar o error en función del algoritmo FIFO
// es una función que se bloquea si no hay procesos listos y se desbloquea sola si llegan a venir nuevos procesos listos
func (f *Fifo) Planificar() (TCB, error) {
	// Bloqueate hasta que alguien te mande algo por este channel -> quién manda por este channel? -> AddToReady()
	// Entonces, bloqueate hasta que alguien agregue un hilo a ready.
	<-PendingThreadsChannel

	var nextTcb *TCB
	var err error

	// Fifo lo único que hace para seleccionar procesos es tomar el primero que entró
	nextTcb, err = f.ready.GetAndRemoveNext()
	if err != nil {
		return TCB{}, errors.New("se quiso obtener un hilo y no habia ningun hilo en ready")
	}

	// Retorná el hilo elegido
	return *nextTcb, nil
}

// AddToReady Le avisa al STS (versión FIFO) que hay un nuevo proceso listo
func (f *Fifo) AddToReady(tcb TCB) error {
	// Agregá el proceso a la cola fifo
	f.ready.Add(&tcb)

	// Mandá mensaje por el canal, o sea, permití que una vuelta más de Planificar() ejecute
	go func() {
		PendingThreadsChannel <- true
	}()
	return nil
}
