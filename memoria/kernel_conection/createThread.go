package kernel_conection

import (
	"bufio"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
)

func CreateThreadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	var request types.RequestToMemory
	body, err := io.ReadAll(r.Body)
	err = json.Unmarshal(body, &request)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Extraer thread y argumentos

	thread := request.Thread
	pid := thread.PID
	tid := thread.TID
	pseudoCodigoAEjecutar := request.Arguments[0]

	// Log obligatorio
	logger.Info("## Hilo Creado - (PID:TID) - (%v,%v)", pid, tid)

	// Guardar el contexto en ExecContext, usando PID y TID como clave
	// La memorySize y base es lo mismo que el del padre
	padreContex := memoriaGlobals.ExecContext[types.Thread{PID: thread.PID}]
	context := types.ExecutionContext{
		MemorySize: padreContex.MemorySize,
		MemoryBase: padreContex.MemoryBase,
	}
	memoriaGlobals.MutexContext.Lock()
	memoriaGlobals.ExecContext[thread] = context
	memoriaGlobals.MutexContext.Unlock()
	logger.Info("Contexto creado para el hilo - (PID:TID): (%v, %v)", pid, tid)

	// Leer el archivo y cargarlo a memoria
	file, err := os.Open(memoriaGlobals.Config.InstructionPath + pseudoCodigoAEjecutar)
	if err != nil {
		logger.Error("No se pudo abrir el archivo de pseudocódigo - %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	if scanner == nil {
		logger.Error("No se pudo crear el scanner")
	}

	for scanner.Scan() {
		instructionRead := scanner.Text()
		if isNotAnInstruction(instructionRead) {
			continue
		}
		memoriaGlobals.CodeRegionForThreads[thread] = append(
			memoriaGlobals.CodeRegionForThreads[thread], instructionRead)
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		logger.Fatal("Error leyendo archivo de pseudocódigo - %v", err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hilo creado correctamente"))
	if err != nil {
		logger.Error("Error escribiendo response: %v", err)
	}
}

func isNotAnInstruction(instruction string) bool {
	return instruction == "" || instruction[0] == '#' || instruction[0] == '\n'
}
