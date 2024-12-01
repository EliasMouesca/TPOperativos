package types

type Tid int
type Pid int

type Thread struct {
	PID Pid `json:"pid"`
	TID Tid `json:"tid"`
}

func (t *Thread) Equals(t2 *Thread) bool {
	return t.PID == t2.PID && t.TID == t2.TID
}
