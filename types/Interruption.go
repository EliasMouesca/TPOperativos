package types

// TODO: Llamar este paquete interrupción en vez de types?

const (
	InterruptionEndOfQuantum = iota
	InterruptionBadInstruction
	InterruptionSegFault
	InterruptionSyscall
)

type Interruption struct {
	Type        int
	Description string
}
