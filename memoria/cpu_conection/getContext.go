package cpu_conection

import (
	"encoding/json"
	global "github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func GetContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	tidS := query.Get("tid")
	pidS := query.Get("pid")

	logger.Debug("Contexto a buscar - (PID:TID) - (%v,%v)", pidS, tidS)

	tid, err := strconv.Atoi(tidS)
	pid, err := strconv.Atoi(pidS)
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	context, exists := global.ExecContext[thread]
	if !exists {
		logger.Error("No se pudo encontrar el contexto")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	logger.Debug("Contexto hayado: %v", context)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(context)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		global.BadRequest(w, r)
		return
	}

	//log obligatorio
	logger.Info("Contexto Solicitado - (PID:TID) - (%v,%v)", pidS, tidS)
}
