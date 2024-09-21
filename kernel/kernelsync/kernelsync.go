package kernelsync

import (
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any, 1)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)

// Sync de Planificador a alrgo plazo
var MutexPlanificadorLP sync.Mutex
var WaitPlanificadorLP sync.WaitGroup

var ChannelProcessArguments = make(chan []string)
var InitProcess = make(chan any)

var ChannelFinishprocess = make(chan int)
var Finishprocess sync.WaitGroup
var SemFinishprocess = make(chan any)
