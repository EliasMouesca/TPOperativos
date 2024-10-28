package cpu_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/helpers"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
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
		logger.Error("Error al obtener los query params")
		http.Error(w, "Error al obtener los query params", http.StatusBadRequest)
	}

	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	// Obtener la instrucción según el PC
	instruccion, err := helpers.GetInstruction(thread, pc)
	if err != nil {
		logger.Error("No se pudo obtener la siguiente linea de código")
		http.Error(w, "No se pudo obtener la siguiente linea de código", http.StatusNotFound)
		return
	}

	// Log obligatorio
	logger.Info("## Obtener instrucción - (PID:TID) - (%v:%v) - Instrucción: %v", pid, tid, instruccion)

	time.Sleep(time.Duration(memoriaGlobals.Config.ResponseDelay))

	// Devolver la instrucción y actualizar el PC
	// Esto no se hace así, porque el que recibe esto no tiene ni puta idea de que recibió,
	// Me voy a tener que ir a fijar directo en tu código como alguna clase de mono?? Jaja chiste, saludos eli
	/*response := struct {
		Instruction string `json:"instruction"`
		PC          int    `json:"pc"`
	}{
		Instruction: instruccion,
		PC:          int(newPC), // El PC avanza para la siguiente instrucción
	}*/

	err = json.NewEncoder(w).Encode(instruccion)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		http.Error(w, "Error al escribir el response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}
