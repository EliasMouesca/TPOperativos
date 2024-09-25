package kernelsync

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any, 1)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)

// MutexPlanificadorLP Sync de Planificador a largo plazo
var MutexPlanificadorLP sync.Mutex
var WaitPlanificadorLP sync.WaitGroup

var ChannelProcessArguments = make(chan []string)
var InitProcess = make(chan any)

var ChannelFinishprocess = make(chan types.Pid)
var Finishprocess sync.WaitGroup
var SemFinishprocess = make(chan any)

var ChannelFinishThread = make(chan int)
var FinishThread sync.WaitGroup
var SemFinishThread = make(chan any)
