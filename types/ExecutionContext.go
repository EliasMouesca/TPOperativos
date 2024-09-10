package types

import (
	"errors"
	"strings"
)

type ExecutionContext struct {
	MemoryBase uint32 `json:"memory_base"`
	MemorySize uint32 `json:"memory_size"`

	Pc uint32 `json:"pc"`
	Ax uint32 `json:"ax"`
	Bx uint32 `json:"bx"`
	Cx uint32 `json:"cx"`
	Dx uint32 `json:"dx"`
	Ex uint32 `json:"ex"`
	Fx uint32 `json:"fx"`
	Gx uint32 `json:"gx"`
	Hx uint32 `json:"hx"`
}

// GetRegister is NOT a getter, this takes a string and returns a reference to the register (if it exists)
func (ectx *ExecutionContext) GetRegister(str string) (*uint32, error) {
	str = strings.ToLower(str)
	switch str {
	case "pc":
		return &ectx.Pc, nil
	case "ax":
		return &ectx.Ax, nil
	case "bx":
		return &ectx.Bx, nil
	case "cx":
		return &ectx.Cx, nil
	case "dx":
		return &ectx.Dx, nil
	case "ex":
		return &ectx.Ex, nil
	case "fx":
		return &ectx.Fx, nil
	case "gx":
		return &ectx.Gx, nil
	case "hx":
		return &ectx.Hx, nil
	default:
		return nil, errors.New("'" + str + "' no constituye ningún registro conocido")
	}
}
