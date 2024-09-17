package kernelsync

import (
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

var PlanificadorLPMutex sync.Mutex
var MemorySemaphore sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any, 1)

var ChannelProcessArguments = make(chan []string)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)

var Wg sync.WaitGroup
