package cpu_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
	"time"
)

func GetInstructionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	tidS := query.Get("tid")
	pidS := query.Get("pid")
	pcS := query.Get("pc")

	tid, err := strconv.Atoi(tidS)
	pid, err := strconv.Atoi(pidS)
	pc, err := strconv.Atoi(pcS)
	if err != nil {
		logger.Error("Erro al obtener las query params")
		http.Error(w, "Erro al obtener las query params", http.StatusBadRequest)
	}
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	instruccion, err := memoria_helpers.GetInstruction(thread, pc)
	if err != nil {
		logger.Error("No se pudo obtener la siguiente linea de código")
		http.Error(w, "No se encontro la instrucción solicitada", http.StatusNotFound)
		return
	}

	// Log obligatorio
	logger.Info("## Obtener instrución - (PID:TID) - (%v:%v) - Instrucción: %v", pid, tid, instruccion)
	time.Sleep(time.Duration(memoria_helpers.Config.ResponseDelay))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(instruccion)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		http.Error(w, "Error al escribir el response", http.StatusInternalServerError)
		return
	}
}
