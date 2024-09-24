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
var BlockedStateQueue types.Queue[kerneltypes.TCB]
var ExitStateQueue types.Queue[kerneltypes.TCB]

var ShortTermScheduler kerneltypes.ShortTermSchedulerInterface

var ExecStateThread kerneltypes.TCB

var Config kerneltypes.KernelConfig

// Map para guardar todos los mutex, dsp cada PCB hace referencia a su id dentro de GlobalMutexRegistry
var GlobalMutexRegistry = map[int]*kerneltypes.MutexWrapper{}
