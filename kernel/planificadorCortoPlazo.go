package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func planificadorCortoPlazo() {
	// Inicializamos el planificador de corto plazo (PCP)
	global.ShortTermScheduler = types.AlgorithmsMap[Config.SchedulerAlgorithm]

	// Mientras vivas, corré lo siguiente
	for {
		// Esta función se bloquea si no hay nada que hacer o si la CPU está ocupada
		tcbToExecute, err := global.ShortTermScheduler.Planificar()
		if err != nil {
			logger.Error("No fue posible planificar cierto hilo - %v", err.Error())
			continue
		}

		// Esperá a que la CPU esté libre / bloqueásela al resto
		global.MutexCPU.Lock()

		// -- A partir de acá tenemos un nuevo proceso en ejecución !! --
		logger.Debug("Hilo a ejecutar: %d", tcbToExecute.TID)

		// TODO: convertir el tcb en Thread

		// TODO: Mandarle el types.Thread a cpu
	}
}
