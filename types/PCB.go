package types

type PCB struct {
	PID   int
	TIDs  []*TCB
	Mutex []int
}
