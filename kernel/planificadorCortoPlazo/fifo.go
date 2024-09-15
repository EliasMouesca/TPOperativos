package planificadorCortoPlazo

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
)

type Fifo struct {
	ready types.Queue[types.TCB]
}

// Planificar devuelve el próximo hilo a ejecutar o error en función del algoritmo FIFO
// es una función que se bloquea si no hay procesos listos y se desbloquea sola si llegan a venir nuevos procesos listos
func (f *Fifo) Planificar() (types.TCB, error) {
	// Si nuestra cola de ready está vacía
	if f.ready.IsEmpty() {
		// Bloqueate hasta que la cola no esté vacía
		global.ReadyQueueNotEmpty.Wait()
	}

	var nextTcb *types.TCB
	var err error

	// Fifo lo único que hace para seleccionar procesos es tomar el primero que entró
	nextTcb, err = f.ready.GetAndRemoveNext()
	if err != nil {
		return types.TCB{}, errors.New("se quiso obtener un hilo y no habia ningun hilo en ready")
	}

	// Retorná el hilo elegido
	return *nextTcb, nil
}

// AddToReady Le avisa al STS (versión FIFO) que hay un nuevo proceso listo
func (f *Fifo) AddToReady(tcb *types.TCB) error {
	// Agregá el proceso a la cola fifo
	f.ready.Add(tcb)

	// Avisá a la función Planificar() que se desbloquee, que hay nuevos procesos.
	global.ReadyQueueNotEmpty.Signal()
	return nil
}
