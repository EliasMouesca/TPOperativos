package ColasMultinivel

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"time"
)

func Max(valor1 time.Duration, duration time.Duration) time.Duration {
	if valor1 > duration {
		return valor1
	} else {
		return duration
	}
}

func EsperarYAvisarFinDeQuantum() error {
	var timer *time.Timer
	tiempoRestante := time.Duration(kernelglobals.Config.Quantum) * time.Millisecond

	for {
		logger.Debug("Inciando round robin, esperando para iniciar quantum")
		<-kernelsync.DebeEmpezarNuevoQuantum
		logger.Info("-- Empieza nuevo Quantum --")
		if len(kernelsync.SyscallChannel) > 0 {
			<-kernelsync.SyscallChannel
			logger.Debug("SyscallChannel consumido: len = %v", len(kernelsync.SyscallChannel))
		}

		timer = time.NewTimer(tiempoRestante)

		select {
		case <-kernelsync.SyscallChannel:
			logger.Debug("Entra a select syscall finalizada")

			if kernelglobals.ExecStateThread != nil {
				logger.Warn("Termina por syscall quantum ignorado")
				if !kernelglobals.ExecStateThread.ExitInstant.After(kernelglobals.ExecStateThread.ExecInstant) {
					logger.Error("No se seteo bien el Exit instant")
				}

				quantumQueLeQuedaba := kernelglobals.ExecStateThread.QuantumRestante
				tiempoEjecutado := kernelglobals.ExecStateThread.ExitInstant.Sub(kernelglobals.ExecStateThread.ExecInstant)

				logger.Debug("Quantum restante previo: %v", quantumQueLeQuedaba)
				logger.Debug("Resta: %v", tiempoEjecutado)

				tiempoRestante = quantumQueLeQuedaba - tiempoEjecutado

				kernelglobals.ExecStateThread.QuantumRestante = Max(tiempoRestante, 0)

				if kernelglobals.ExecStateThread.QuantumRestante == 0 {
					tiempoRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
					enviarFinDeQuantumACPU(tiempoRestante)
				}

				logger.Debug("Tiempo restante de quantum: %v", tiempoRestante)
				logger.Debug("Exit: %v", kernelglobals.ExecStateThread.ExitInstant)
				logger.Debug("Exec: %v", kernelglobals.ExecStateThread.ExecInstant)
			}

		case <-timer.C:
			logger.Warn("Quantum completado, enviando Interrupcion a CPU por fin de quantum")
			if kernelglobals.ExecStateThread != nil {
				tiempoRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
				enviarFinDeQuantumACPU(tiempoRestante)
			}
		}
	}
}

func enviarFinDeQuantumACPU(tiempoRestante time.Duration) {
	tiempoRestante = time.Duration(kernelglobals.Config.Quantum) * time.Millisecond
	kernelglobals.ExecStateThread.QuantumRestante = tiempoRestante
	logger.Debug("Nuevo quantum seteado: %v", tiempoRestante)

	err := shorttermscheduler.CpuInterrupt(
		types.Interruption{
			Type:        types.InterruptionEndOfQuantum,
			Description: "Interrupcion por fin de Quantum",
		})
	if err != nil {
		logger.Error("Error al interrupir a la CPU (fin de quantum) - %v", err)
	}
}
