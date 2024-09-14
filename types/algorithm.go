package types

type Algorithm interface {
	planificar()
	addToReady(tcb TCB)
}
