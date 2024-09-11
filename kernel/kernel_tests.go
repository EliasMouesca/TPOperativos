package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// Estructura para simular el request JSON
type CreateProcessRequest struct {
	Pseudocodigo string `json:"pseudocodigo"`
	ProcessSize  int    `json:"processSize"`
	Prioridad    int    `json:"prioridad"`
}

// Función de test para `PROCESS_CREATE`
func TestProcessCreate(t *testing.T) {
	// Preparar datos de prueba
	requestData := CreateProcessRequest{
		Pseudocodigo: "test_process",
		ProcessSize:  1024,
		Prioridad:    1,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		t.Fatalf("Error al serializar el cuerpo de la solicitud: %v", err)
	}

	// Crear un request simulado
	req, err := http.NewRequest("POST", "/kernel/createProcess", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Error al crear el request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

}
