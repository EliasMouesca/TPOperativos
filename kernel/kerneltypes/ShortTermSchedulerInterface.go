package kerneltypes

import "github.com/sisoputnfrba/tp-golang/types"

type ShortTermSchedulerInterface interface {
	Planificar() (*TCB, error)
	AddToReady(*TCB) error
	ThreadExists(types.Tid, types.Pid) (bool, error)
	ThreadRemove(types.Tid, types.Pid) error
	Init()
}
