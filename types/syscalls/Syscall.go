package syscalls

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

func New(syscallType int, description string, arguments []string) Syscall {
	return Syscall{
		Type:        syscallType,
		Description: description,
		Arguments:   arguments,
	}
}

var SyscallNames = map[int]string{
	DumpMemory:    "DumpMemory",
	IO:            "IO",
	ProcessCreate: "ProcessCreate",
	ThreadCreate:  "ThreadCreate",
	ThreadJoin:    "ThreadJoin",
	ThreadCancel:  "ThreadCancel",
	MutexCreate:   "MutexCreate",
	MutexLock:     "MutexLock",
	MutexUnlock:   "MutexUnlock",
	ThreadExit:    "ThreadExit",
	ProcessExit:   "ProcessExit",
}
