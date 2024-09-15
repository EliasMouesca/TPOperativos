package shorttermscheduler

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
)

type ColasMultiNivel struct {
	readyQueue []*types.Queue[kerneltypes.TCB]
}

func (cmm *ColasMultiNivel) Planificar() (kerneltypes.TCB, error) {
	<-kernelsync.PendingThreadsChannel

	nextTcb, err := cmm.getNextTcb()
	if err != nil {
		return nextTcb, err
	}

	return nextTcb, nil
}

func (cmm *ColasMultiNivel) AddToReady(tcb kerneltypes.TCB) error {
	// Inicializo la cola si es la primera vez que se llama
	if cmm.readyQueue == nil {
		cmm.readyQueue = make([]*types.Queue[kerneltypes.TCB], 0)
	}

	inserted := false
	for i := range cmm.readyQueue {
		// Verifico si ya existe una cola de la prioridad del hilo
		if cmm.readyQueue[i].Priority == tcb.Prioridad {
			// Si existe lo agrego de forma FIFO a la cola y salgo
			cmm.readyQueue[i].Add(&tcb)
			inserted = true
			break
		}
	}

	// Si no existe una lista de esa prioridad
	if !inserted {
		err := cmm.addNewQueue(tcb)
		if err != nil {
			return err
		}
	}

	go func() {
		kernelsync.PendingThreadsChannel <- true
	}()

	return nil
}

func (cmm *ColasMultiNivel) addNewQueue(tcb kerneltypes.TCB) error {
	// Creo la cola y la agrego al slice de colas
	newQueue := new(types.Queue[kerneltypes.TCB])
	newQueue.Priority = tcb.Prioridad
	newQueue.Add(&tcb)

	// Buscar la posición correcta para insertar la nueva cola
	insertedAt := false
	for i := range cmm.readyQueue {
		if newQueue.Priority < cmm.readyQueue[i].Priority {
			// Insertar la nueva cola en la posición `i` sin remover otros elementos
			cmm.readyQueue = append(cmm.readyQueue[:i], append([]*types.Queue[kerneltypes.TCB]{newQueue}, cmm.readyQueue[i:]...)...)
			insertedAt = true
			break
		}
	}
	// Si la prioridad es la menor (número más alto), se agrega al final
	if !insertedAt {
		cmm.readyQueue = append(cmm.readyQueue, newQueue)
	}

	return nil
}

func (cmm *ColasMultiNivel) getNextTcb() (kerneltypes.TCB, error) {
	for i := range cmm.readyQueue {
		if !cmm.readyQueue[i].IsEmpty() {
			nextTcb, err := roundRobin(cmm.readyQueue[i])
			if err != nil {
				return kerneltypes.TCB{}, err
			}
			return nextTcb, nil
		}
	}
	return kerneltypes.TCB{}, errors.New("se quizo hacer un getNextTcb y no habia ningun tcb en ready")
}

func roundRobin(queue *types.Queue[kerneltypes.TCB]) (kerneltypes.TCB, error) {
	// TODO: implementar
	return kerneltypes.TCB{}, nil
}
