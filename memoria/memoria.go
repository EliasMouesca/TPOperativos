package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}
}

func main() {
	logger.Info("--- Comienzo ejecución MEMORIA ---")

	// --- COMUNICACIONES ---
	GenerateRequest("cpu", "8080")
	GenerateRequest("kernel", "8081")

	// --- INICIALIZAR EL SERVER ---

	hostname := "localhost"
	port := "8082"

	http.HandleFunc("/memoria/accion", GoodRequest) // Si el POST es /memoria/accion => redirecciona a una GoodRequest
	http.HandleFunc("/", BadRequest)                // Si el POST es / => redirecciona a una BadRequest

	logger.Info("Server activo en %v:%v", hostname, port)
	err := http.ListenAndServe(hostname+":"+port, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}

	// --- FIN DE INICIALIZACION DE SERVER ---

}

// --- REQUESTS ---

type BodyRequest struct {
	Message string `json:"message"`
	Origin  string `json:"origin"`
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonError, err := json.MarshalIndent("Bad Request", "", "  ")
	_, err = w.Write(jsonError)
	if err != nil {
		logger.Error("Error: se ha recibido una URL incorrecta - BadRequest")
	}
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
