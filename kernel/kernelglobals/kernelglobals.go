package kernelglobals

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
)

// NewStateQueue Colas de New, Blocked y Exit usando el tipo Queue, quedaria cambiar donde se usan
var NewStateQueue types.Queue[*kerneltypes.PCB]
var BlockedStateQueue types.Queue[*kerneltypes.TCB]
var ExitStateQueue types.Queue[*kerneltypes.TCB]

var ShortTermScheduler kerneltypes.ShortTermSchedulerInterface

var EveryTCBInTheKernel []kerneltypes.TCB
var EveryPCBInTheKernel []kerneltypes.PCB

var ExecStateThread *kerneltypes.TCB

var Config kerneltypes.KernelConfig
