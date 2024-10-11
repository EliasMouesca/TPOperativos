package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

var mutexNames = []string{"mutex_A", "mutex_B", "mutex_C", "mutex_D"}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func generateRandomName(base string) string {
	randomNumber := rand.Intn(100) // Genera un número aleatorio entre 0 y 99
	return base + "_" + string(rune(randomNumber))
}

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

// Test para crear un proceso con nombre aleatorio
func TestProcessCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	processName := generateRandomName("test_process")

	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessCreate,
		Arguments: []string{processName, "1024", "1"},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ProcessCreate syscall para proceso %s enviado correctamente.", processName)
}

// Test para crear un hilo con nombre aleatorio
func TestThreadCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	threadName := generateRandomName("thread_code")

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCreate,
		Arguments: []string{threadName, "1"},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ThreadCreate syscall enviado correctamente para %s.", threadName)
}

// Test para cancelar un hilo
func TestThreadCancelKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCancel,
		Arguments: []string{"3"},
	}
	sendSyscallRequest(t, syscall)
	t.Log("ThreadCancel syscall enviado correctamente.")
}

// Test para ThreadJoin
func TestThreadJoinKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Vamos a suponer que queremos joinear el TID 2
	tidToJoin := "2"

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadJoin,
		Arguments: []string{tidToJoin},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ThreadJoin syscall enviado correctamente para el TID %s.", tidToJoin)
}

// Test para crear un mutex con nombre aleatorio
func TestMutexCreateKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mutexName := generateRandomName("mutex")

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexCreate,
		Arguments: []string{mutexName},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexCreate syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para hacer lock a un mutex con nombre aleatorio
func TestMutexLockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mutexName := generateRandomName("mutex")

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: []string{mutexName},
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexLock syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para desbloquear un mutex con nombre aleatorio
func TestMutexUnlockKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	mutexName := generateRandomName("mutex")

	syscall := syscalls.Syscall{
		Type:      syscalls.MutexUnlock,
		Arguments: []string{mutexName}, // Desbloquear el mutex que se eligió
	}
	sendSyscallRequest(t, syscall)
	t.Logf("MutexUnlock syscall enviado correctamente (Nombre: %s).", mutexName)
}

// Test para finalizar un proceso (ProcessExit) basado en el PID del proceso en ejecución
func TestProcessExitKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Aquí asumo que el kernel mantiene el PID del proceso en ejecución
	pid := "1" // Puedes modificar esto para tomar el PID dinámicamente

	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessExit,
		Arguments: []string{pid}, // PID del proceso a finalizar
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ProcessExit syscall enviado correctamente para el proceso con PID %s.", pid)
}

func TestThreadExitKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadExit,
		Arguments: []string{}, // PID del proceso a finalizar
	}
	sendSyscallRequest(t, syscall)
	t.Logf("ThreadExit syscall enviado correctamente.")
}

// Test para DumpMemory
func TestDumpMemoryKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// No necesitamos argumentos adicionales para este caso
	syscall := syscalls.Syscall{
		Type:      syscalls.DumpMemory,
		Arguments: []string{}, // Sin argumentos
	}
	sendSyscallRequest(t, syscall)
	t.Log("DumpMemory syscall enviada correctamente.")
}

// Test para IO
func TestIOKERNEL(t *testing.T) {
	time.Sleep(2 * time.Second)

	// Vamos a suponer que queremos bloquear el hilo por 500 milisegundos
	blockedTime := "500"

	syscall := syscalls.Syscall{
		Type:      syscalls.IO,
		Arguments: []string{blockedTime}, // Tiempo de bloqueo en milisegundos
	}
	sendSyscallRequest(t, syscall)
	t.Logf("IO syscall enviada correctamente, con un tiempo de bloqueo de %s milisegundos.", blockedTime)
}
