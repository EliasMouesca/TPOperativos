package main

import "errors"

type CpuConfig struct {
	SelfAddress   string `json:"ip_self"`
	SelfPort      int    `json:"port_self"`
	MemoryAddress string `json:"ip_memory"`
	MemoryPort    int    `json:"port_memory"`
	KernelAddress string `json:"ip_kernel"`
	KernelPort    int    `json:"port_kernel"`
	LogLevel      string `json:"log_level"`
}

func (cfg CpuConfig) validate() error {
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

	if cfg.LogLevel == "" {
		return errors.New("falta el campo 'log_level'")
	}
	return nil
}
