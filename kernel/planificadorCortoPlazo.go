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
		logger.Trace("Empezando nueva planificación")
		logCurrentState("DESPUËS DE EMPEZAR A PLANIFICAR")
		logger.Trace("Length PendingThreadsChannel %v", len(kernelsync.PendingThreadsChannel))

		// Bloqueate hasta que alguien te mande algo por este channel -> quién manda por este channel? -> AddToReady()
		// Entonces, bloqueate hasta que alguien agregue un hilo a ready.

		// Bloqueate si hay una syscall en progreso, no queremos estar ejecutando a la vez que la syscall
		<-kernelsync.SyscallFinalizada
		logger.Trace("No hay una syscall activa o finalizó, planificando")

		var tcbToExecute *kerneltypes.TCB
		var err error

		if kernelglobals.ExecStateThread != nil {
			tcbToExecute = kernelglobals.ExecStateThread
		} else {
			<-kernelsync.PendingThreadsChannel
			logger.Trace("Hay hilos en ready para planificar")

			tcbToExecute, err = kernelglobals.ShortTermScheduler.Planificar()
			if err != nil {
				logger.Error("No fue posible planificar cierto hilo - %v", err.Error())
				continue
			}
		}

		logger.Trace("Tratando de lockear la CPU para enviar nuevo proceso")
		// Esperá a que la CPU esté libre / bloqueásela al resto
		kernelsync.MutexCPU.Lock()
		logger.Trace("CPU Lockeada, mandando a execute")

		// -- A partir de acá tenemos un nuevo proceso en ejecución !! --

		//Crafteo proximo hilo
		nextThread := types.Thread{TID: tcbToExecute.TID, PID: tcbToExecute.FatherPCB.PID}
		data, err := json.Marshal(nextThread)

		// Envio proximo hilo a cpu
		url := fmt.Sprintf("http://%v:%v/cpu/execute", kernelglobals.Config.CpuAddress, kernelglobals.Config.CpuPort)
		_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
		if err != nil {
			logger.Error("Error en request")
		}
		kernelglobals.ExecStateThread = tcbToExecute

		logger.Debug("## (<%v>:<%v>) Ejecutando hilo", tcbToExecute.FatherPCB.PID, tcbToExecute.TID)

		// TODO: Qué es esto?
		go func() {
			kernelsync.QuantumChannel <- time.After(time.Duration(kernelglobals.Config.Quantum) * time.Millisecond)
		}()

		logger.Trace("Finalizó la planificación")

	}
}
