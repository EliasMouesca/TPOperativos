package types

import CortoPlazo "github.com/sisoputnfrba/tp-golang/kernel/planificadorCortoPlazo"

type ShortTermScheduler interface {
	Planificar() (TCB, error)
	AddToReady(tcb *TCB) error
}

var AlgorithmsMap = map[string]ShortTermScheduler{
	"FIFO": CortoPlazo.Fifo{},
	"P":    CortoPlazo.Prioridades{},
	"CMM":  CortoPlazo.CMM{},
}
