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

// Test para crear 4 procesos
func TestProcessCreateKERNEL(t *testing.T) {
	// Esperar un tiempo para asegurarse de que el servidor kernel esté activo
	time.Sleep(2 * time.Second)

	for i := 0; i < 2; i++ {
		syscall := syscalls.Syscall{
			Type:      syscalls.ProcessCreate,
			Arguments: []string{"test_process", "1024", "1"},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("ProcessCreate syscall #%d enviado correctamente.", i+1)
	}
}

// Test para crear 5 hilos para el proceso del hilo en ejecución
func TestThreadCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	for i := 0; i < 5; i++ {
		syscall := syscalls.Syscall{
			Type:      syscalls.ThreadCreate,
			Arguments: []string{"thread_code", "1"}, // Cambia los argumentos según el pseudocódigo y la prioridad del hilo
		}
		sendSyscallRequest(t, syscall)
		t.Logf("ThreadCreate syscall #%d enviado correctamente.", i+1)
	}
}

// Test para cancelar un hilo
func TestThreadCancelKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCancel,
		Arguments: []string{"3"}, // TID del hilo a cancelar
	}
	sendSyscallRequest(t, syscall)
	t.Log("ThreadCancel syscall enviado correctamente.")
}

// Test para crear 4 mutex
func TestMutexCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	for i := 0; i < 4; i++ {
		mutexName := "mutex_" + string(rune('A'+i)) // Mutex_A, Mutex_B, etc.
		syscall := syscalls.Syscall{
			Type:      syscalls.MutexCreate,
			Arguments: []string{mutexName},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("MutexCreate syscall #%d enviado correctamente (Nombre: %s).", i+1, mutexName)
	}
}

// Test para hacer lock a 2 mutex
func TestMutexLockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	for i := 0; i < 2; i++ {
		mutexName := "mutex_" + string(rune('A'+i)) // Bloquear Mutex_A, Mutex_B
		syscall := syscalls.Syscall{
			Type:      syscalls.MutexLock,
			Arguments: []string{mutexName},
		}
		sendSyscallRequest(t, syscall)
		t.Logf("MutexLock syscall #%d enviado correctamente (Nombre: %s).", i+1, mutexName)
	}
}

// Test para desbloquear un mutex
func TestMutexUnlockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexUnlock,
		Arguments: []string{"mutex_A"}, // Desbloquear Mutex_A
	}
	sendSyscallRequest(t, syscall)
	t.Log("MutexUnlock syscall enviado correctamente (Nombre: mutex_A).")
}

// Test para finalizar un proceso (ProcessExit)
func TestProcessExitKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Supongamos que queremos finalizar el proceso con PID 1
	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessExit,
		Arguments: []string{"1"}, // PID del proceso a finalizar
	}
	sendSyscallRequest(t, syscall)
	t.Log("ProcessExit syscall enviado correctamente para el proceso con PID 1.")
}
