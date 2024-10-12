package kernelsync

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
)

// MutexCPU la cpu es una sola -> mutex
var MutexCPU sync.Mutex

// PendingThreadsChannel un canal para que el corto sepa que hay procesos pendientes de planificar
var PendingThreadsChannel = make(chan any)

// QuantumChannel se manda una señal por acá cuando se acabó el quantum
var QuantumChannel = make(chan any)

// MutexPlanificadorLP Sync de Planificador a largo plazo
var MutexPlanificadorLP sync.Mutex
var WaitPlanificadorLP sync.WaitGroup

var ChannelProcessArguments = make(chan []string)
var InitProcess = make(chan any)
var SemProcessCreateOK = make(chan struct{}, 1)

var ChannelFinishprocess = make(chan types.Pid)

var ChannelFinishThread = make(chan []string, 1)
var ThreadExitComplete = make(chan struct{}, 1)

var ChannelThreadCreate = make(chan []string)
var ThreadCreateComplete = make(chan struct{}, 1)

var ChannelIO = make(chan *kerneltypes.TCB)
var ChannelIO2 = make(chan int)
var SemIo = make(chan int)

// SyscallFinalizada La idea es: no pongas otro proceso a ejecutar si la syscall que llamaron no terminó !!
// en una PC de verdad no tenés una CPU para syscalls y otra para proces, xd
var SyscallFinalizada = make(chan any)
