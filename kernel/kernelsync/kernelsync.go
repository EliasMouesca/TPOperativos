package kernelsync

import (
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

var MutexPlanificadorLP sync.Mutex
var MemorySemaphore sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any, 1)

var ChannelProcessArguments = make(chan []string)
var SemCreateprocess = make(chan any)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)

var WaitPlanificadorLP sync.WaitGroup
var InitProcess sync.WaitGroup
