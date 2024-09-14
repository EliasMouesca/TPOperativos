package global

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
)

//var NEW []types.PCB
//var Ready []types.TCB

// Colas de New y Ready usnaod el tipo Queue, quedaria cambiar donde se usan
var NEW types.Queue[types.PCB]
var Ready types.Queue[types.TCB]

var ShortTermScheduler types.ShortTermScheduler
var MutexCPU sync.Mutex

var ChanReady = make(chan any, 1)
