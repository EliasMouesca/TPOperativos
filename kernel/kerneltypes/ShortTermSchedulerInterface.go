package kerneltypes

import "github.com/sisoputnfrba/tp-golang/types"

type ShortTermSchedulerInterface interface {
	Planificar() (*TCB, error)
	AddToReady(*TCB) error
	ThreadExists(types.Thread) (bool, error) // Si existe -> true, sino -> false
	ThreadRemove(types.Thread) error         // SACA DE READY (si existe, sino, error)
}
