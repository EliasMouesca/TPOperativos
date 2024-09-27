package types

import "errors"

// En este archivo se encuentran las estructuras
// de las request que le envia el Planificador de Largo Plazo
// a memoria

// TODO: tipo de direcciones a memoria

// la direccion en la cual esta la handleFunc de memoria
// por ejemplo: http.HandleFunc("/kernel/createProcess", createProcess)
const (
	CreateProcess = "createProcess"
	FinishProcess = "finishProcess"
	CreateThread  = "createThread"
	FinishThread  = "finishThread"
	MemoryDump    = "memoryDump"
)

type RequestToMemory struct {
	Type      string   `json:"type"`
	Arguments []string `json:"arguments"`
}

var ErrorRequestType = map[string]error{
	CreateProcess: errors.New("Memoria: No hay espacio disponible en memoria "),
	FinishProcess: errors.New("Memoria: No se puedo finalizar el proceso"),
	CreateThread:  errors.New("Memoria: No se puedo crear el hilo"),
	FinishThread:  errors.New("Memoria: No se pudo finalizar el hilo"),
	// DumpMemory:
}
