package types

type Tid int
type Pid int

type Thread struct {
	PID Pid `json:"pid"`
	TID Tid `json:"tid"`
}
