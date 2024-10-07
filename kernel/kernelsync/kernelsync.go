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
var SemFinishprocess = make(chan any)

var ChannelFinishThread = make(chan []string, 1)
var SemFinishThread = make(chan struct{}, 1)
var SemMovedFinishThreads = make(chan struct{}, 1)

var ChannelThreadCreate = make(chan []string)
var SemThreadCreate = make(chan any)
