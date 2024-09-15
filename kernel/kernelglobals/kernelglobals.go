package kernelglobals

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
)

//var NEW []types.PCB
//var Ready []types.TCB

// Colas de New y Ready usnaod el tipo Queue, quedaria cambiar donde se usan
var NewStateQueue types.Queue[kerneltypes.PCB]
var ReadyStateQueue types.Queue[kerneltypes.TCB]

var ShortTermScheduler kerneltypes.ShortTermSchedulerInterface

var ExecStateThread kerneltypes.TCB
