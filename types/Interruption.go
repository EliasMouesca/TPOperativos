package types

const (
	EndOfQuantum = iota
	BadInstruction
	Segfault
	Syscall
)

type Interruption struct {
	Type        int
	Description string
}
