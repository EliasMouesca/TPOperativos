package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/ColasMultinivel"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Prioridades"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"time"
)

var AlgorithmsMap = map[string]kerneltypes.ShortTermSchedulerInterface{
	"FIFO": &Fifo.Fifo{},
	"P":    &Prioridades.Prioridades{},
	"CMN":  &ColasMultinivel.ColasMultiNivel{},
}

func planificadorCortoPlazo() {
	kernelglobals.ShortTermScheduler.Init()
	// Mientras vivas, corré lo siguiente
	for {
		logger.Trace("Empezando nueva planificación")
		logCurrentState("DESPUÉS DE EMPEZAR A PLANIFICAR")
		logger.Trace("Length PendingThreadsChannel %v", len(kernelsync.PendingThreadsChannel))

		// Bloqueate si hay una syscall en progreso, no queremos estar ejecutando a la vez que la syscall
		logger.Debug("Esperando que termine una syscall")
		<-kernelsync.SyscallFinalizada
		logger.Trace("No hay una syscall activa o finalizó, planificando")

		var tcbToExecute *kerneltypes.TCB
		var err error

		kernelsync.MutexExecThread.Lock()
		if kernelglobals.ExecStateThread != nil {
			logger.Trace("ExecStateThread: (PID: %v - TID: %v)", kernelglobals.ExecStateThread.FatherPCB.PID, kernelglobals.ExecStateThread.TID)
			tcbToExecute = kernelglobals.ExecStateThread
			kernelsync.MutexExecThread.Unlock()

			tcbToExecute.HayQuantumRestante = true

			logger.Debug("Devuelvo hilo sin planificar")
		} else {
			kernelsync.MutexExecThread.Unlock()
			logger.Debug("Esperando que haya hilos en ready")
			<-kernelsync.PendingThreadsChannel
			logger.Trace("Hay hilos en ready para planificar")

			tcbToExecute, err = kernelglobals.ShortTermScheduler.Planificar()
			logger.Debug("Hilo a planificar (<%v>:<%v>)", tcbToExecute.FatherPCB.PID, tcbToExecute.TID)

			tcbToExecute.HayQuantumRestante = false

			if err != nil {
				logger.Error("No fue posible planificar cierto hilo - %v", err.Error())
				continue
			}
		}

		// Esperá a que la CPU esté libre / bloqueásela al resto
		logger.Trace("Tratando de lockear la CPU para enviar nuevo proceso")
		kernelsync.MutexCPU.Lock()
		logger.Trace("CPU Lockeada, mandando a execute")

		// -- A partir de acá tenemos un nuevo proceso en ejecución !! --

		//Crafteo proximo hilo
		nextThread := types.Thread{TID: tcbToExecute.TID, PID: tcbToExecute.FatherPCB.PID}

		logger.Debug("Enviando TCB a CPU")
		// Envio proximo hilo a cpu
		shorttermscheduler.CpuExecute(nextThread)

		kernelglobals.ExecStateThread = tcbToExecute
		tcbToExecute.ExecInstant = time.Now()

		logger.Debug("Asinando nuevo hilo a ExecStateThread: (TID: %v PID: %v)", tcbToExecute.TID, tcbToExecute.FatherPCB.PID)
		if kernelglobals.Config.SchedulerAlgorithm == "CMN" {
			go func() {
				logger.Debug("Mandamos que debe empezar nuevo quantum")
				kernelsync.DebeEmpezarNuevoQuantum <- true
			}()
		}

		logger.Debug("## (<%v>:<%v>) Ejecutando hilo", tcbToExecute.FatherPCB.PID, tcbToExecute.TID)
		go func() {
			logger.Debug("Antes de mandar true a channel de planifterminada")
			kernelsync.PlanificacionFinalizada <- true
			logger.Debug("Despues de mandar true a channel de planifterminada")
		}()

		logger.Trace("Finalizó la planificación")
	}
}
