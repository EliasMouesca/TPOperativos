package types

// TODO: O este paquete se llama syscall o al enum le prefijamos "syscall", DumpMemory -> SyscallDumpMemory

// Types of syscalls
const (
	DumpMemory = iota
	IO
	ProcessCreate
	ThreadCreate
	ThreadJoin
	ThreadCancel
	MutexCreate
	MutexLock
	MutexUnlock
	ThreadExit
	ProcessExit
)

type Syscall struct {
	Type        int
	Description string
	Arguments   []string
}
