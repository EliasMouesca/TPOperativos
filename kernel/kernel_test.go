package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"net/http"
	"testing"
	"time"
)

func sendSyscallRequest(t *testing.T, syscall syscalls.Syscall) {
	// Serializar la syscall en JSON
	jsonData, err := json.Marshal(syscall)
	if err != nil {
		t.Fatalf("Error al serializar la syscall: %v", err)
	}

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
}

func TestKernelSyscalls(t *testing.T) {
	// Esperar un tiempo para asegurarse de que el servidor kernel esté activo
	time.Sleep(2 * time.Second)

	// 1. Crear 4 procesos
	for i := 0; i < 4; i++ {
		syscall := syscalls.Syscall{
			Type:      syscalls.ProcessCreate,
			Arguments: []string{"test_process", "1024", "1"},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("ProcessCreate syscall #%d enviado correctamente.", i+1)
	}

	// 2. Hacer 5 ThreadCreate para el proceso del hilo ejecutando
	for i := 0; i < 5; i++ {
		syscall := syscalls.Syscall{
			Type:      syscalls.ThreadCreate,
			Arguments: []string{"thread_code", "1"}, // Cambia los argumentos según el pseudocódigo y la prioridad del hilo
		}
		sendSyscallRequest(t, syscall)
		t.Logf("ThreadCreate syscall #%d enviado correctamente.", i+1)
	}

	// 3. Hacer 1 ThreadCancel a un hilo de los creados
	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCancel,
		Arguments: []string{"3"},
	}
	sendSyscallRequest(t, syscall)
	t.Log("ThreadCancel syscall enviado correctamente.")

	// 4. Hacer 4 MutexCreate
	for i := 0; i < 4; i++ {
		mutexName := "mutex_" + string(rune('A'+i)) // Mutex_A, Mutex_B, etc.
		syscall := syscalls.Syscall{
			Type:      syscalls.MutexCreate,
			Arguments: []string{mutexName},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("MutexCreate syscall #%d enviado correctamente (Nombre: %s).", i+1, mutexName)
	}

	// 5. Hacer 2 MutexLock
	for i := 0; i < 2; i++ {
		mutexName := "mutex_" + string(rune('A'+i)) // Bloquear Mutex_A, Mutex_B
		syscall := syscalls.Syscall{
			Type:      syscalls.MutexLock,
			Arguments: []string{mutexName},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("MutexLock syscall #%d enviado correctamente (Nombre: %s).", i+1, mutexName)
	}

	// 6. Hacer 1 MutexUnlock
	syscall = syscalls.Syscall{
		Type:      syscalls.MutexUnlock,
		Arguments: []string{"mutex_A"}, // Desbloquear Mutex_A
	}
	sendSyscallRequest(t, syscall)
	t.Log("MutexUnlock syscall enviado correctamente (Nombre: mutex_A).")
}
