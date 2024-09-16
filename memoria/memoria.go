package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"os"
)

// TODAVÍA NO FUNCIONA, EL LUNES/MARTES LO TERMINO Y HAGO QUE FUNCIONES (vos confía)

var config MemoriaConfig
var contextFile *os.File
var indexContext map[types.Thread]int64

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
		os.Exit(1)
	}

	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	if err = config.validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

	//Create contextsFile from zero
	contextFile, err = os.OpenFile(config.ContextsFileName, os.O_TRUNC, 0666) //Don't know what 0666 means
	if err != nil {
		logger.Fatal("No se pudo crear el archivo: %v" + config.ContextsFileName)
	}
	defer func() {
		if err := contextFile.Close(); err != nil {
			logger.Fatal("Error al cerrar el archivo:", err)
		}
	}()
	logger.Debug("Archivo de contextos creado exitosamente")
}

func main() {
	logger.Info("--- Comienzo ejecución MEMORIA ---")

	// --- INICIALIZAR EL SERVER ---
	// TODO: -----------------
	http.HandleFunc("POST /memoria/getContext", GetContext)
	http.HandleFunc("POST /memoria/saveContext", saveContext)
	http.HandleFunc("POST /memoria/updateExecutionContext", GoodRequest)
	http.HandleFunc("POST /memoria/getInstruction", GoodRequest)
	http.HandleFunc("POST /memoria/ReadMem", GoodRequest)
	http.HandleFunc("POST /WriteMem", GoodRequest)
	// -----------------------
	http.HandleFunc("/", BadRequest)

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Info("Server activo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}
	// --- FIN DE INICIALIZACION DE SERVER ---
}

func GetContext(w http.ResponseWriter, r *http.Request) {
	//Request log
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, r)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var thread types.Thread
	err = json.Unmarshal(body, &thread)
	if err != nil {
		BadRequest(w, r)
		return
	}

	contextFile, err = os.OpenFile(config.ContextsFileName, os.O_RDONLY, 0666)
	if err != nil {
		logger.Error("No se pudo acceder al archivo: %v" + config.ContextsFileName)
		return
	}

	//Place cursor at start of thread context
	_, err = contextFile.Seek(getContextAboslutePosition(thread), 0)
	if err != nil {
		logger.Error("No se pudo localizar el contexto de (tid: %v, pid: %v) dentro del archivo", thread.Tid, thread.Pid)
		return
	}

	var context types.ExecutionContext
	err = binary.Read(contextFile, binary.LittleEndian, &context)
	if err != nil {
		logger.Error("No se pudo leer el archivo: %v" + config.ContextsFileName)
		return
	}

	//Close contexts file
	defer func() {
		if err := contextFile.Close(); err != nil {
			logger.Fatal("Error al cerrar el archivo:", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func saveContext(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		BadRequest(w, r)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var contexto types.ExecutionContext
	err = json.Unmarshal(body, &contexto)
	if err != nil {
		BadRequest(w, r)
		return
	}

	contextFile, err = os.OpenFile(config.ContextsFileName, os.O_RDWR, 0666)
	if err != nil {
		logger.Error("No se pudo acceder al archivo: %v" + config.ContextsFileName)
		return
	}

	posicion, exists := indexContext[contexto.Thread]
	// If context to be saved doesn't exist => save at end of file
	if !exists {
		_, err = contextFile.Seek(0, 2)
		indexContext[contexto.Thread], _ = contextFile.Seek(0, 0) //Save position into index
	} else {
		_, err = contextFile.Seek(posicion, 0)
	}
	err = binary.Write(contextFile, binary.LittleEndian, contexto) //Save context into file
	if err != nil {
		logger.Error("Error al modificar el archivo de contextos")
	}
	logger.Debug("Archivo de contextos modificado exitosamente")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Contexto guardado exitosamente"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err.Error())
	}
}

// Done in separate function for possible future changes
func getContextAboslutePosition(thread types.Thread) int64 {
	return indexContext[thread]
}

// --- REQUESTS ---
func BadRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request inválida: %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request mal formada"))
	if err != nil {
		logger.Error("Error al escribir el response a %v", r.RemoteAddr)
	}
}

// BORRAR

type BodyRequest struct {
	Message string `json:"message"`
	Origin  string `json:"origin"`
}

func GenerateRequest(receiver string, port string) {
	body := BodyRequest{
		Message: "Hola " + receiver,
		Origin:  "memoria",
	}
	bodyJson, err := json.Marshal(body)
	if err != nil {
		logger.Error("Error al serializar JSON: ", err)
	}
	response, err := http.Post("http://localhost:"+port+"/"+receiver+"/accion", "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		logger.Error("Error al hacer la request")
	} else {
		logger.Info("Respuesta exitosa de %s: %v", receiver, response.Status)
	}
}

// -- END REQUESTS --

// --- RESPONSE --

type Response struct {
	Response string      `json:"response"`
	Request  BodyRequest `json:"request"`
}

func GoodRequest(w http.ResponseWriter, r *http.Request) {
	var request BodyRequest
	if r.Body != nil {
		requestBody, err := io.ReadAll(r.Body)
		err = json.Unmarshal(requestBody, &request)
		if err != nil {
			logger.Error("Error al leer la request")
		}
	}
	response := Response{
		Request:  request,
		Response: "Solicitud recibida de " + request.Origin,
	}
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		logger.Error("Error al craftear la respuesta")
	}

	logger.Info("Hola " + request.Origin + "! Respuesta exitosa.")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		logger.Error("Error al responder a la request")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}
