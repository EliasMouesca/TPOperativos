package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/global"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

func planificadorCortoPlazo() {
	global.ShortTermScheduler = types.AlgorithmsMap[Config.SchedulerAlgorithm]
	for {
		global.MutexCPU.Lock() // Hace unlock en la API que expone kernel
		// El if tiene que ser otro sem wait Ready
		tcbToExecute, err := global.ShortTermScheduler.Planificar()
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		// TODO: convertir el tcb en Thread y mandar a cpu
		logger.Debug("Hilo a ejecutar: %d", tcbToExecute.TID)
	}
}
