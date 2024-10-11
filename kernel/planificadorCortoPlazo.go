package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/ColasMultinivel"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Prioridades"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"time"
)

var AlgorithmsMap = map[string]kerneltypes.ShortTermSchedulerInterface{
	"FIFO": &Fifo.Fifo{},
	"P":    &Prioridades.Prioridades{},
	"CMM":  &ColasMultinivel.ColasMultiNivel{},
}

func planificadorCortoPlazo() {
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

		// Crafteo proximo hilo
		nextThread := types.Thread{TID: tcbToExecute.TID, PID: tcbToExecute.FatherPCB.PID}
		data, err := json.Marshal(nextThread)

		// Envio proximo hilo a cpu
		url := fmt.Sprintf("http://%v:%v/cpu/execute", kernelglobals.Config.CpuAddress, kernelglobals.Config.CpuPort)
		_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			return
		}
		kernelsync.QuantumChannel <- time.After(time.Duration(kernelglobals.Config.Quantum) * time.Millisecond)
		kernelglobals.ExecStateThread = tcbToExecute

		logCurrentState("DESPUES DE PLANIFICAR")
	}
}
