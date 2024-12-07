package cpu_conection

import (
	"encoding/binary"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/helpers"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
	"time"
)

func ReadMemoryHandler(w http.ResponseWriter, r *http.Request) {
	defer time.Sleep(time.Millisecond * time.Duration(memoriaGlobals.Config.ResponseDelay))

	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("addr")
	tidS := query.Get("tid")
	pidS := query.Get("pid")

	// Que es el Tamaño????????
	// Log obligatorio
	logger.Info("## Lectura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS, "")

	dir, err := strconv.Atoi(dirS)
	if err != nil {
		logger.Error("Dirección física mal formada: %v", dirS)
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
		return
	}

	cautroMordidas, err := helpers.ReadMemory(dir)
	if err != nil {
		logger.Error("Error al leer la dirección: %v", dir)
		http.Error(w, "No se pudo leer la dirección de memoria", http.StatusNotFound)
		return
	}

	data := binary.BigEndian.Uint32(cautroMordidas[:])

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		http.Error(w, "Error al escribir el response", http.StatusInternalServerError)
		return
	}
}
