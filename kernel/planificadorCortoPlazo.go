package main

import (
	CortoPlazo "github.com/sisoputnfrba/tp-golang/kernel/planificadorCortoPlazo"
	"github.com/sisoputnfrba/tp-golang/types"
)

var algorithm string

func planificadorCortoPlazo() {
	algorithm := Config.SchedulerAlgorithm

	switch algorithm {
	case "FIFO":
		CortoPlazo.Fifo()
	case "P":
		CortoPlazo.Prioridades()
	case "CMM":
		CortoPlazo.ColasMultiNivel()
	}
}

func AddToReady(tcb types.TCB) {
	switch algorithm {
	case "FIFO":
		CortoPlazo.AddToFifo(tcb)
	case "P":
		CortoPlazo.AddToP(tcb)
	case "CMM":
		CortoPlazo.AddToCMM(tcb)
	}
}
