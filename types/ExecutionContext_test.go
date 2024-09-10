package types

import "testing"

func TestGetRegister(t *testing.T) {
	context := ExecutionContext{}
	addr, err := context.GetRegister("ax")
	if err != nil {
		t.Error(err)
	}
	if &context.Ax != addr {
		t.Error("Get register is not working for ax")
	}

}
