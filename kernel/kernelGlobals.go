package main

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
)

//var NEW []types.PCB
//var Ready []types.TCB

// Colas de New y Ready usnaod el tipo Queue, quedaria cambiar donde se usan
var NEW types.Queue[PCB]
var Ready types.Queue[TCB]
var EXIT = make(chan TCB, 1) // Canal con capacidad de 1

var ShortTermScheduler ShortTermSchedulerInterface
var MutexCPU sync.Mutex

var PendingThreadsChannel = make(chan any, 1)

var currentThread TCB
