package cpu_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
	"time"
)

func GetContextHandler(w http.ResponseWriter, r *http.Request) {
	defer time.Sleep(time.Millisecond * time.Duration(memoriaGlobals.Config.ResponseDelay))

	logger.Trace("Memoria entró a GetContextHandler()")
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
	if err != nil {
		logger.Error("No se pudo obtener los query params")
		http.Error(w, "No se pudo obtener los query params", http.StatusBadRequest)
		return
	}
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	context, exists := memoriaGlobals.ExecContext[thread]
	if !exists {
		logger.Error("No se pudo encontrar el contexto")
		http.Error(w, "No se pudo encontrar el contexto", http.StatusNotFound)
		return
	}

	//context := types.ExecutionContext{}

	logger.Debug("Contexto hayado: %v", context)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(context)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		http.Error(w, "Error al encontrar el contexto", http.StatusInternalServerError)
		return
	}

	//log obligatorio
	logger.Info("Contexto Solicitado - (PID:TID) - (%v,%v)", pidS, tidS)
}
