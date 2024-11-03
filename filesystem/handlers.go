package main

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"log"
	"net/http"
	"os"
)

type MemoryDumpRequest struct {
	Nombre    string `json:"Nombre"`
	Size      int    `json:"Size"`
	Contenido []byte `json:"Contenido"`
}

func persistMemoryDump(w http.ResponseWriter, r *http.Request) {
	logger.Debug("MemoryDump solicitado")
	var dumpRequest MemoryDumpRequest
	err := json.NewDecoder(r.Body).Decode(&dumpRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error al decodificar el cuerpo de la solicitud:", err)
		w.Write([]byte("Solicitud inválida: Error al decodificar JSON"))
		return
	}

	file, err := os.Create(dumpRequest.Nombre)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error creando archivo:", dumpRequest.Nombre, "-", err)
		w.Write([]byte("Error al crear el archivo"))
		return
	}
	defer file.Close()

	_, err = file.Write(dumpRequest.Contenido)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Error escribiendo contenido en archivo:", dumpRequest.Nombre, "-", err)
		w.Write([]byte("Error al escribir en el archivo"))
		return
	}

	// Responder con éxito
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo bien, dump persistido :)"))

	logger.Debug("MemoryDump completado")

	//TODO: Eli querido, te comentamos todo porque recibias query y a nosotros no nos gusta.
	// Ademas, te lo dejamos andando, denada :)

	//pidS := r.URL.Query().Get("pid")
	//tidS := r.URL.Query().Get("tid")
	//sizeS := r.URL.Query().Get("size")
	//
	//if pidS == "" || tidS == "" || sizeS == "" {
	//	w.WriteHeader(http.StatusBadRequest)
	//	logger.Error("Request inválida recibida")
	//	w.Write([]byte("No se recibieron parámetros adecuados. Lee filesystem.md"))
	//	return
	//}
	//
	//pid, err := strconv.Atoi(pidS)
	//tid, err := strconv.Atoi(tidS)
	//size, err := strconv.Atoi(sizeS)
	//if err != nil {
	//	w.WriteHeader(http.StatusBadRequest)
	//	logger.Error("Request inválida recibida")
	//	w.Write([]byte("El query param pid, tid o size no se pudo convertir a un número"))
	//	return
	//
	//}
	//
	//binData, err := io.ReadAll(r.Body)
	//if err != nil {
	//	w.WriteHeader(http.StatusInternalServerError)
	//	logger.Error("No se pudo leer el body de la request")
	//	w.Write([]byte("Server error!!"))
	//	return
	//}
	//
	//var data []byte
	//json.Unmarshal(binData, &data)
	//
	///*
	//	filename := fmt.Sprintf("%d-%d-%s.dmp", pid, tid, time.Now().Format("20060102-150405"))
	//	file, err := os.Create(filename)
	//	if err != nil {
	//		logger.Error("Error creando archivo '%v' - %v", filename, err)
	//	}
	//
	//	file.Write()
	//*/
	//
	//w.Write([]byte("Todo bien, dump persistido :)"))
	//w.WriteHeader(http.StatusOK)
	//return
}
