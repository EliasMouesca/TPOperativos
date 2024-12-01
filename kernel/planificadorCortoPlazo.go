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
	"CMN":  &ColasMultinivel.ColasMultiNivel{},
}

func planificadorCortoPlazo() {
	// Mientras vivas, corré lo siguiente
	for {
		logger.Trace("Empezando nueva planificación")
		logCurrentState("DESPUÉS DE EMPEZAR A PLANIFICAR")
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
			logger.Trace("DEVUELVO HILO SIN PLANIFICAR!")
		} else {
			<-kernelsync.PendingThreadsChannel
			logger.Trace("Hay hilos en ready para planificar")

			tcbToExecute, err = kernelglobals.ShortTermScheduler.Planificar()
			logger.Debug("Hilo a planificar (<%v>:<%v>)", tcbToExecute.FatherPCB.PID, tcbToExecute.TID)

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
		logger.Debug("tcbToExecute: %v", tcbToExecute.TID)
		kernelglobals.ExecStateThread = tcbToExecute

		logger.Debug("## (<%v>:<%v>) Ejecutando hilo", tcbToExecute.FatherPCB.PID, tcbToExecute.TID)

		if kernelglobals.Config.SchedulerAlgorithm == "CMN" {
			go func() {
				timer := time.NewTimer(time.Duration(kernelglobals.Config.Quantum) * time.Millisecond)

				select {
				case <-timer.C:
					// El temporizador expiró, envía al canal
					if tcbToExecute == kernelglobals.ExecStateThread {
						logger.Debug("Fin de quantum!")
						kernelsync.QuantumChannel <- struct{}{} // Enviar un valor vacío al canal
					}
				}
			}()
		}
		go func() {
			logger.Debug("Antes de mandar true a channel de planifterminada")
			kernelsync.PlanificacionFinalizada <- true
			logger.Debug("Despues de mandar true a channel de planifterminada")
		}()
		logger.Trace("Finalizó la planificación")
	}
}
