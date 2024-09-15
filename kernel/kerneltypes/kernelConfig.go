package kerneltypes

import (
	"errors"
)

type KernelConfig struct {
	SelfAddress        string `json:"ip_self"`
	SelfPort           int    `json:"port_self"`
	MemoryAddress      string `json:"ip_memory"`
	MemoryPort         int    `json:"port_memory"`
	CpuAddress         string `json:"ip_cpu"`
	CpuPort            int    `json:"port_cpu"`
	SchedulerAlgorithm string `json:"scheduler_algorithm"`
	Quantum            int    `json:"quantum"`
	LogLevel           string `json:"log_level"`
}

func (cfg KernelConfig) Validate() error {
	if cfg.SelfAddress == "" {
		return errors.New("falta el campo 'ip_self'")
	}
	if cfg.SelfPort == 0 {
		return errors.New("falta el campo 'port_self' o es inválido")
	}
	if cfg.MemoryAddress == "" {
		return errors.New("falta el campo 'ip_memory'")
	}
	if cfg.MemoryPort == 0 {
		return errors.New("falta el campo 'port_memory' o es inválido")
	}
	if cfg.CpuAddress == "" {
		return errors.New("falta el campo 'ip_cpu'")
	}
	if cfg.CpuPort == 0 {
		return errors.New("falta el campo 'port_cpu' o es inválido")
	}
	if cfg.SchedulerAlgorithm == "" {
		return errors.New("falta el campo 'scheduler_algorithm'")
	}
	if cfg.Quantum == 0 {
		return errors.New("falta el campo 'quantum' o es inválido")
	}
	if cfg.LogLevel == "" {
		return errors.New("falta el campo 'log_level'")
	}
	return nil
}
