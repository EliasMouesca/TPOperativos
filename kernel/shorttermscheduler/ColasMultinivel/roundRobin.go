package ColasMultinivel

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"time"
)

func EsperarYAvisarFinDeQuantum() error {
	for {
		logger.Debug("Inciando round robin, esperando para iniciar quantum")
		<-kernelsync.DebeEmpezarNuevoQuantum
		logger.Info("-- Empieza nuevo Quantum --")
		if len(kernelsync.SyscallChannel) > 0 {
			logger.Warn("El SyscallChannel tenia una entrada sin consumir")
			<-kernelsync.SyscallChannel
			logger.Debug("SyscallChannel consumido: len = %v", len(kernelsync.SyscallChannel))
		}

		select {
		case <-kernelsync.SyscallChannel:
			logger.Warn("Termina por syscall quantum cancelado")

		case <-time.NewTimer(time.Duration(kernelglobals.Config.Quantum) * time.Millisecond).C:
			logger.Warn("Quantum completado")
			logger.Debug("Enviando Interrupcion a CPU por fin de quantum")

			err := shorttermscheduler.CpuInterrupt(
				types.Interruption{
					Type:        types.InterruptionEndOfQuantum,
					Description: "Interrupcion por fin de Quantum",
				})
			if err != nil {
				logger.Error("Error al interrupir a la CPU (fin de quantum) - %v", err)
				return err
			}
		}
	}
}
