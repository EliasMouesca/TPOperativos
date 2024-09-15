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
var EXIT = make(chan types.TCB, 1) // Canal con capacidad de 1

var ShortTermScheduler types.ShortTermScheduler
var MutexCPU sync.Mutex

// Este mutex solo se usa para inicializar el sync.Cond, no usar para otra cosa
var readyQueueMutex sync.Mutex

// ReadyQueueNotEmpty es un semáforo que avisa que la cola de procesos listos ya no está vacía
// TODO quizás esto no es general y debería estar adentro de cada algo de planif. corto plazo ?
var ReadyQueueNotEmpty *sync.Cond = sync.NewCond(&readyQueueMutex)
