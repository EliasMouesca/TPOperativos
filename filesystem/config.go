package main

import "errors"

type fsConfig struct {
	SelfAddress      string `json:"ip_self"`
	SelfPort         int    `json:"port_self"`
	MemoryAddress    string `json:"ip_memory"`
	MemoryPort       int    `json:"port_memory"`
	MountDir         string `json:"mount_dir"`
	BlockSize        int    `json:"block_size"`
	BlockCount       int    `json:"block_count"`
	BlockAccessDelay int    `json:"block_access_delay"`
	LogLevel         string `json:"log_level"`
}

func (cfg fsConfig) validate() error {
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

	if cfg.MountDir == "" {
		return errors.New("falta el campo 'mount_dir'")
	}

	if cfg.BlockSize <= 0 {
		return errors.New("falta el campo 'block_size' o es inválido")
	}

	if cfg.BlockCount <= 0 {
		return errors.New("falta el campo 'block_count' o es inválido")
	}

	if cfg.BlockAccessDelay < 0 {
		return errors.New("falta el campo 'block_access_delay' o es inválido")
	}

	if cfg.LogLevel == "" {
		return errors.New("falta el campo 'log_level'")
	}

	return nil
}
