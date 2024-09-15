package kerneltypes

type PCB struct {
	PID   int
	TIDs  []int
	Mutex []int
}
