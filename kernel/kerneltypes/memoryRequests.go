package kerneltypes

import "errors"

// tipo de direcciones a memoria
const (
	CreateProcess = "createProcess"
	FinishProcess = "finishProcess"
	CreateThread  = "createThread"
	FinishThread  = "finishThread"
	MemoryDump    = "memoryDump"
)

type MemoryRequest struct {
	Type      string   `json:"type"`
	Arguments []string `json:"arguments"`
}

var MapErrorRequestType = map[string]error{
	CreateProcess: errors.New("No hay espacio disponible en memoria "),
	// FinishProcess:
	// CreateThread:
	// FinishThread:
	// DumpMemory:
}
