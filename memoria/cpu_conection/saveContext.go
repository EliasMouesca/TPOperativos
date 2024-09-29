package cpu_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
	"time"
)

func SaveContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("# Request recibida de: %v", r.RemoteAddr)
	// Get pid and tid from query params
	queryParams := r.URL.Query()
	tidS := queryParams.Get("tid")
	pidS := queryParams.Get("pid")

	logger.Trace("Contexto a guardar - (PID:TID) - (%v,%v)", pidS, tidS)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("No se pudo leer el cuerpo del request")
		http.Error(w, "No se pudo leer el cuerpo del request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var contexto types.ExecutionContext
	err = json.Unmarshal(body, &contexto)
	if err != nil {
		logger.Error("No se pudo decodificar el cuerpo del request")
		http.Error(w, "No se decodificar el cuerpo del request", http.StatusBadRequest)
		return
	}

	tid, err := strconv.Atoi(tidS)
	pid, err := strconv.Atoi(pidS)
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	_, exists := memoria_helpers.ExecContext[thread]
	if !exists {
		logger.Trace("No existe el thread buscado, se creará un nuevo contexto")
	}
	memoria_helpers.ExecContext[thread] = contexto
	logger.Debug("Contexto guardado exitosamente: %v", memoria_helpers.ExecContext[thread])

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Contexto guardado exitosamente"))
	if err != nil {
		logger.Error("Error escribiendo el response - %v", err.Error())
		http.Error(w, "Error escribiendo el response", http.StatusInternalServerError)
		return
	}

	// Log obligatorio
	logger.Info("## Contexto Actualizado - (PID:TID) - (%v:%v)", pidS, tidS)
	time.Sleep(time.Duration(memoria_helpers.Config.ResponseDelay))
}
