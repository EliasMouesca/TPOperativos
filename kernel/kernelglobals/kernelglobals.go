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

// TODO: Qué es esto? -eli
var EXIT = make(chan kerneltypes.TCB, 1) // Canal con capacidad de 1

var ShortTermScheduler kerneltypes.ShortTermSchedulerInterface

var currentThread kerneltypes.TCB
