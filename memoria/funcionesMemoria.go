package main

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
)

func dump(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	pidS := query.Get("pid")
	tidS := query.Get("tid")

	// Log obligatorio
	logger.Info("## Memory Dump solicitado - (PID:TID) - (%v:%v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
}

func finishThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	pidS := query.Get("pid")
	tidS := query.Get("tid")

	// Log obligatorio
	logger.Info("## Hilo Destruido - (PID:TID) - (%v,%v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
}

func createThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	pidS := query.Get("pid")
	tidS := query.Get("tid")

	// Log obligatorio
	logger.Info("## Hilo Creado - (PID:TID) - (%v,%v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
}

func finishProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	sizeS := query.Get("size")
	pidS := query.Get("ptd")

	//Log obligatorio
	logger.Info("## Proceso Destruido -  PID: %v - Tamaño: %v", pidS, sizeS)

	w.WriteHeader(http.StatusOK)
}

func createProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	sizeS := query.Get("size")
	pidS := query.Get("ptd")

	//Log obligatorio
	logger.Info("## Proceso Creado -  PID: %v - Tamaño: %v", pidS, sizeS)

	w.WriteHeader(http.StatusOK)
}

func writeMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("dir")
	tidS := query.Get("tid")
	pidS := query.Get("pid")
	// Log obligatorio
	logger.Info("## Escritura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS)

	dir, err := strconv.Atoi(dirS)
	if err != nil {
		logger.Error("Dirección física mal formada: %v", dirS)
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("No se pudo leer el cuerpo del request")
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var data [4]byte
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("No se pudo decodificar el cuerpo del request")
		BadRequest(w, r)
		return
	}

	err = writeMemoryPosta(dir, data)
	if err != nil {
		logger.Error("Error al escribir en memoria de usuario")
		BadRequest(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func readMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("dir")
	tidS := query.Get("tid")
	pidS := query.Get("pid")

	// Que es el Tamaño????????
	// Log obligatorio
	logger.Info("## Lectura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS)

	dir, err := strconv.Atoi(dirS)
	if err != nil {
		logger.Error("Dirección física mal formada: %v", dirS)
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
	}

	data, err := readMemoryPosta(dir)
	if err != nil {
		logger.Error("Error al leer la dirección: %v", dir)
		http.Error(w, "No se pudo leer la dirección de memoria", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		BadRequest(w, r)
		return
	}
}

func getInstruction(w http.ResponseWriter, r *http.Request) {
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
	thread := types.Thread{PID: types.Pid(pid), TID: types.Tid(tid)}

	instruccion, err := getInstructionPosta(thread, pc)
	if err != nil {
		logger.Error("No se pudo obtener la siguiente linea de código")
		http.Error(w, "No se encontro la instrucción solicitada", http.StatusNotFound)
		return
	}

	// Log obligatorio
	logger.Info("## Obtener instrución - (PID:TID) - (%v:%v) - Instrucción: %v", pid, tid, instruccion)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(instruccion)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		BadRequest(w, r)
		return
	}
}

func saveContext(w http.ResponseWriter, r *http.Request) {
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
		BadRequest(w, r)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var contexto types.ExecutionContext
	err = json.Unmarshal(body, &contexto)
	if err != nil {
		logger.Error("No se pudo decodificar el cuerpo del request")
		BadRequest(w, r)
		return
	}

	tid, err := strconv.Atoi(tidS)
	pid, err := strconv.Atoi(pidS)
	thread := types.Thread{PID: pid, TID: tid}

	_, exists := execContext[thread]
	if !exists {
		logger.Trace("No existe el thread buscado, se creará un nuevo contexto")
	}
	execContext[thread] = contexto
	logger.Debug("Contexto guardado exitosamente: %v", execContext[thread])

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Contexto guardado exitosamente"))
	if err != nil {
		logger.Error("Error escribiendo el response - %v", err.Error())
		BadRequest(w, r)
	}

	// Log obligatorio
	logger.Info("## Contexto Actualizado - (PID:TID) - (%v:%v)", pidS, tidS)
}

func getContext(w http.ResponseWriter, r *http.Request) {
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
	thread := types.Thread{PID: pid, TID: tid}

	context, exists := execContext[thread]
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
		BadRequest(w, r)
		return
	}

	//log obligatorio
	logger.Info("Contexto Solicitado - (PID:TID) - (%v,%v)", pidS, tidS)
}
