package main

import (
	"bytes"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

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
		logger.Error("Error al serializar JSON - %v", err)
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
