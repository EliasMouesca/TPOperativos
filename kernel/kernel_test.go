package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"net/http"
	"sync"
	"testing"
	"time"
)

var (
	processCounter int
	threadCounter  int
	mutexCounter   int
	mutexNames     = []string{"mutex_A", "mutex_B", "mutex_C", "mutex_D"}
	mu             sync.Mutex
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

// Test para crear un proceso
func TestProcessCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Usar un mutex para asegurar que el contador se incremente correctamente entre ejecuciones paralelas
	mu.Lock()
	processCounter++
	processName := "test_process_" + string(rune(processCounter))
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessCreate,
		Arguments: []string{processName, "1024", "1"},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ProcessCreate syscall para proceso %s enviado correctamente.", processName)
}

// Test para crear un hilo para el proceso del hilo en ejecución
func TestThreadCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mu.Lock()
	threadCounter++
	threadName := "thread_code_" + string(rune(threadCounter))
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCreate,
		Arguments: []string{threadName, "1"}, // Cambia los argumentos según el pseudocódigo y la prioridad del hilo
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ThreadCreate syscall enviado correctamente para %s.", threadName)
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

// Test para crear un mutex
func TestMutexCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mu.Lock()
	mutexName := mutexNames[mutexCounter%len(mutexNames)] // Usar el nombre basado en el contador
	mutexCounter++
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexCreate,
		Arguments: []string{mutexName},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexCreate syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para hacer lock a un mutex
func TestMutexLockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mu.Lock()
	mutexName := mutexNames[mutexCounter%len(mutexNames)]
	mutexCounter++
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: []string{mutexName},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexLock syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para desbloquear un mutex
func TestMutexUnlockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mu.Lock()
	mutexName := mutexNames[mutexCounter%len(mutexNames)]
	mutexCounter++
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexUnlock,
		Arguments: []string{mutexName}, // Desbloquear el mutex que se eligió
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexUnlock syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para finalizar un proceso (ProcessExit)
func TestProcessExitKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Suponiendo que queremos finalizar el proceso más reciente (PID basado en el processCounter)
	mu.Lock()
	pid := processCounter
	mu.Unlock()

	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessExit,
		Arguments: []string{string(rune(pid))}, // PID del proceso a finalizar
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ProcessExit syscall enviado correctamente para el proceso con PID %d.", pid)
}
