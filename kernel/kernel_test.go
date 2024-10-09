package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"net/http"
	"testing"
	"time"
)

func TestKernelSyscall(t *testing.T) {
	// Crear un ejemplo de syscall (puedes cambiar el tipo y los argumentos según tu caso)
	syscall := syscalls.Syscall{
		Type:      2,                                     // Tipo de syscall que quieres probar
		Arguments: []string{"test_process", "1024", "1"}, // Argumentos para la syscall
	}

	// Serializar la syscall en JSON
	jsonData, err := json.Marshal(syscall)
	if err != nil {
		t.Fatalf("Error al serializar la syscall: %v", err)
	}

	// Esperar un tiempo para asegurarse de que el servidor kernel esté activo
	time.Sleep(2 * time.Second)

	// Enviar la solicitud POST al servidor del kernel
	resp, err := http.Post("http://127.0.0.1:8081/kernel/syscall", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Error al enviar la syscall al kernel: %v", err)
	}
	defer resp.Body.Close()

	// Verificar si la respuesta es correcta
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Respuesta inesperada del kernel, estado HTTP: %d", resp.StatusCode)
	}

	// Aquí puedes agregar más verificaciones sobre el comportamiento del kernel si es necesario.
	t.Log("Test de syscall ProcessCreate enviado y recibido correctamente.")
}
