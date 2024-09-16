package kernelsync

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

var PlanificadorLPMutex sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any, 1)

var ChannelProcessCreate = make(chan kerneltypes.PCB)
var ChannelProcessArguments = make(chan []string)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)
