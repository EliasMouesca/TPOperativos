package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

var AlgorithmsMap = map[string]kerneltypes.ShortTermSchedulerInterface{
	"FIFO": &shorttermscheduler.Fifo{},
	"P":    &shorttermscheduler.Prioridades{},
	"CMM":  &shorttermscheduler.ColasMultiNivel{},
}

func planificadorCortoPlazo() {
	// Inicializamos el planificador de corto plazo (PCP)
	kernelglobals.ShortTermScheduler = AlgorithmsMap[Config.SchedulerAlgorithm]

	// Mientras vivas, corré lo siguiente
	for {
		// Esta función se bloquea si no hay nada que hacer o si la CPU está ocupada
		tcbToExecute, err := kernelglobals.ShortTermScheduler.Planificar()
		if err != nil {
			logger.Error("No fue posible planificar cierto hilo - %v", err.Error())
			continue
		}

		// Esperá a que la CPU esté libre / bloqueásela al resto
		kernelsync.MutexCPU.Lock()

		// -- A partir de acá tenemos un nuevo proceso en ejecución !! --
		logger.Debug("Hilo a ejecutar: %d", tcbToExecute.TID)

		// TODO: convertir el tcb en Thread

		// TODO: Mandarle el types.Thread a cpu
	}
}
