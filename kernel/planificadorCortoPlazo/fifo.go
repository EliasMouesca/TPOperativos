package planificadorCortoPlazo

import (
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

var Ready types.Queue[types.TCB]

func Planificar() {
	logger.Debug("--- Comienzo ejecucion de FIFO ---")
	for {
		if global.Ready.IsEmpty() {
			logger.Debug("No hay hilos en Ready")
		} else {
			nextTcb := global.Ready.GetAndRemoveNext()
			logger.Debug("Hilo a ejecutar: %d", nextTcb.TID)
		}
	}
}

func AddToReady(tcb *types.TCB) {
	Ready.Add(tcb)
}
