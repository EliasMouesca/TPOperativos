package planificadorCortoPlazo

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Fifo struct {
	ready types.Queue[types.TCB]
}

func (f *Fifo) Planificar() (types.TCB, error) {
	logger.Debug("Llamada a planificacion fifo")
	var nextTcb *types.TCB
	var err error

	nextTcb, err = f.ready.GetAndRemoveNext()
	if err != nil {
		return types.TCB{}, errors.New("se quiso obtener un hilo y no habia ningun hilo en ready")
	}

	return *nextTcb, nil
}

func (f *Fifo) AddToReady(tcb *types.TCB) error {
	f.ready.Add(tcb)

	return nil
}
