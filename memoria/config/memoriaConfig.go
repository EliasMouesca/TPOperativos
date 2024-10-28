package config

type MemoriaConfig struct {
	SelfPort        int    `json:"self_port"`
	SelfAddress     string `json:"self_address"`
	MemorySize      int    `json:"memory_size"`
	InstructionPath string `json:"instruction_path"`
	ResponseDelay   int    `json:"response_delay"`
	IpKernel        string `json:"ip_kernel"`
	PortKernel      int    `json:"port_kernel"`
	IpCpu           string `json:"ip_cpu"`
	PortCpu         int    `json:"port_cpu"`
	IpFilesystem    string `json:"ip_filesystem"`
	PortFilesystem  int    `json:"port_filesystem"`
	Scheme          string `json:"scheme"`
	SearchAlgorithm string `json:"search_algorithm"`
	Partitions      []int  `json:"partitions"`
	LogLevel        string `json:"log_level"`
}

// TODO
func (cfg MemoriaConfig) Validate() error {
	return nil
}
