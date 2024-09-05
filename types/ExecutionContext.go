package types

type ExecutionContext struct {
	MemoryBase int `json:"memory_base"`
	MemorySize int `json:"memory_size"`

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
