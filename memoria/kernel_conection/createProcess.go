package kernel_conection

import (
	"bufio"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
	"strconv"
)

func CreateProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Leer el cuerpo de la request
	var requestData types.RequestToMemory
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	// Extraer el PID y el tamaño desde el cuerpo JSON
	pid := requestData.Thread.PID
	mainThread := types.Thread{
		PID: pid,
		TID: types.Tid(0),
	}
	size, _ := strconv.Atoi(requestData.Arguments[1])
	pseudoCodigoAEjecutar := requestData.Arguments[0]

	if size%4 != 0 {
		logger.Error("El tamaño del proceso no es multiplo de 4")
		http.Error(w, "El tamaño del proceso no es multiplo de 4", http.StatusBadRequest)
		return
	}

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
		memoriaGlobals.CodeRegionForThreads[mainThread] = append(
			memoriaGlobals.CodeRegionForThreads[mainThread], instructionRead)
	}

	base, err := memoriaGlobals.SistemaParticiones.AsignarProcesoAParticion(types.Pid(pid), size)
	if err != nil {
		logger.Warn("No se pudo asignar el proceso < %v > de tamaño %v a una particion de memoria", pid, size)
		if err.Error() == types.Compactacion {
			logger.Debug("Se debe compactar")
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Trace("MemoryBase: %v; MemorySize: %v", base, size)
	context := types.ExecutionContext{}
	context.MemoryBase = base
	context.MemorySize = uint32(size)

	memoriaGlobals.MutexContext.Lock()
	memoriaGlobals.ExecContext[mainThread] = context
	memoriaGlobals.MutexContext.Unlock()

	logger.Info("Contexto creado para el hilo - (PID:TID): (%v, %v): %v", pid, 0, context)

	// Log obligatorio
	logger.Info("## Proceso Creado - PID: < %v > - Tamaño: < %v >", pid, size)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Proceso creado correctamente"))
}
