package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

type SyscallRequest struct {
	Type        int      `json:"type"`
	Arguments   []string `json:"arguments"`
	Description string   `json:"description"`
}

func main() {
	logger.Info("--- Comienzo ejecución KERNEL-TEST ---")

	// --- COMUNICACIONES ---
	logger.Info("Enviando syscall para crear proceso")
	GenerateSyscallRequest("kernel", "8081", createProcessSyscall("test_process_1", 1024, 1))

	logger.Info("Enviando syscall para crear un hilo")
	GenerateSyscallRequest("kernel", "8081", createThreadSyscall(1, "thread_1", 1))

	logger.Info("Enviando syscall para finalizar un hilo")
	GenerateSyscallRequest("kernel", "8081", threadExitSyscall(1, 0))

	logger.Info("Enviando syscall para finalizar un proceso")
	GenerateSyscallRequest("kernel", "8081", processExitSyscall(1))

	// --- INICIALIZAR EL SERVER ---

	hostname := "localhost"
	port := "8090"

	http.HandleFunc("/", BadRequest)

	logger.Info("Server activo en %v:%v", hostname, port)
	err := http.ListenAndServe(hostname+":"+port, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8090: " + err.Error())
	}

	// --- FIN DE INICIALIZACION DE SERVER ---
}

// --- REQUESTS ---

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonError, err := json.MarshalIndent("Bad Request", "", "  ")
	_, err = w.Write(jsonError)
	if err != nil {
		logger.Error("Error: se ha recibido una URL incorrecta - BadRequest")
	}
}

func GenerateSyscallRequest(receiver string, port string, syscallReq SyscallRequest) {
	bodyJson, err := json.Marshal(syscallReq)
	if err != nil {
		logger.Error("Error al serializar JSON: ", err)
	}
	response, err := http.Post("http://localhost:"+port+"/"+receiver+"/syscall", "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		logger.Error("Error al hacer la request")
	} else {
		logger.Info("Respuesta exitosa de %s: %v", receiver, response.Status)
	}
}

// --- FUNCIONES PARA GENERAR LAS SYSCALLS ---

func createProcessSyscall(processName string, processSize int, prioridad int) SyscallRequest {
	return SyscallRequest{
		Type:        syscalls.ProcessCreate,
		Arguments:   []string{processName, strconv.Itoa(processSize), strconv.Itoa(prioridad)},
		Description: "Create Process",
	}
}

func createThreadSyscall(pid int, threadName string, prioridad int) SyscallRequest {
	return SyscallRequest{
		Type:        syscalls.ThreadCreate,
		Arguments:   []string{strconv.Itoa(pid), threadName, strconv.Itoa(prioridad)},
		Description: "Create Thread",
	}
}

func processExitSyscall(pid int) SyscallRequest {
	return SyscallRequest{
		Type:        syscalls.ProcessExit,
		Arguments:   []string{strconv.Itoa(pid)},
		Description: "Process Exit",
	}
}

func threadExitSyscall(pid int, tid int) SyscallRequest {
	return SyscallRequest{
		Type:        syscalls.ThreadExit,
		Arguments:   []string{strconv.Itoa(pid), strconv.Itoa(tid)},
		Description: "Thread Exit",
	}
}
