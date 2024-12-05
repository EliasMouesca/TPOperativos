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

		if kernelglobals.ExecStateThread != nil {
			tcbToExecute = kernelglobals.ExecStateThread
			//go func() {
			//
			//	kernelglobals.QuantumTimer.Stop()
			//	kernelsync.DebeEmpezarNuevoQuantum <- true
			//}()
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

		// Esperá a que la CPU esté libre / bloqueásela al resto
		logger.Trace("Tratando de lockear la CPU para enviar nuevo proceso")
		kernelsync.MutexCPU.Lock()
		logger.Trace("CPU Lockeada, mandando a execute")

		// -- A partir de acá tenemos un nuevo proceso en ejecución !! --

		//Crafteo proximo hilo
		nextThread := types.Thread{TID: tcbToExecute.TID, PID: tcbToExecute.FatherPCB.PID}
		//data, err := json.Marshal(nextThread)

		logger.Debug("Enviando TCB a CPU")
		// Envio proximo hilo a cpu
		shorttermscheduler.CpuExecute(nextThread)
		//url := fmt.Sprintf("http://%v:%v/cpu/execute", kernelglobals.Config.CpuAddress, kernelglobals.Config.CpuPort)
		//_, err = http.Post(url, "application/json", bytes.NewBuffer(data))
		//if err != nil {
		//	logger.Error("Error en request")
		//}
		//logger.Debug("tcbToExecute: %v", tcbToExecute.TID)

		kernelglobals.ExecStateThread = tcbToExecute
		logger.Debug("Asinando nuevo hilo a ExecStateThread: (TID: %v)", tcbToExecute.TID)
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
