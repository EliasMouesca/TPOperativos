package planificadorCortoPlazo

import (
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func Fifo() {
	logger.Debug("--- Comienzo ejecucion de FIFO ---")

	if global.Ready.isEmpty {
		logger.Debug("No hay hilos en Ready")
	} else {
		nextTcb := global.Ready.getAndRemoveNext()
		logger.Debug("Hilo a ejecutar: %d", nextTcb.TID)
	}
}
