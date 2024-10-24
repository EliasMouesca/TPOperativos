package main

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
)

func persistMemoryDump(w http.ResponseWriter, r *http.Request) {
	pidS := r.URL.Query().Get("pid")
	tidS := r.URL.Query().Get("tid")
	sizeS := r.URL.Query().Get("size")

	if pidS == "" || tidS == "" || sizeS == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Request inválida recibida")
		w.Write([]byte("No se recibieron parámetros adecuados. Lee filesystem.md"))
		return
	}

	pid, err := strconv.Atoi(pidS)
	tid, err := strconv.Atoi(tidS)
	size, err := strconv.Atoi(sizeS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Request inválida recibida")
		w.Write([]byte("El query param pid, tid o size no se pudo convertir a un número"))
		return

	}

	binData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("No se pudo leer el body de la request")
		w.Write([]byte("Server error!!"))
		return
	}

	var data []byte
	json.Unmarshal(binData, &data)

	/*
		filename := fmt.Sprintf("%d-%d-%s.dmp", pid, tid, time.Now().Format("20060102-150405"))
		file, err := os.Create(filename)
		if err != nil {
			logger.Error("Error creando archivo '%v' - %v", filename, err)
		}

		file.Write()
	*/

	w.Write([]byte("Todo bien, dump persistido :)"))
	w.WriteHeader(http.StatusOK)
	return
}
