package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

type Response struct {
	Response string      `json:"response"`
	Request  BodyRequest `json:"request"`
}
type BodyRequest struct {
	Message string `json:"message"`
	Origin  string `json:"origin"`
}

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("cpu.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}
}

func main() {
	logger.Info("--- Comienzo ejecución CPU ---")

	http.HandleFunc("POST /cpu/accion", GenerateSendResponse)
	http.HandleFunc("/", BadRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8080: " + err.Error())
	}

	GenerateRequest("kernel")
	GenerateRequest("memoria")
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonError, err := json.MarshalIndent("Bad Request", "", "  ")
	_, err = w.Write(jsonError)
	if err != nil {
		logger.Error("Error al escribir la respuesta a BadRequest")
	}
}

func GenerateSendResponse(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		logger.Error("Error al responder a la request")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}

func GenerateRequest(receiver string) {
	body := BodyRequest{
		Message: "Hola " + receiver,
		Origin:  "Cpu",
	}
	bodyJson, err := json.MarshalIndent(body, "", "  ")
	response, err := http.Post("POST", "http://localhost:8080/"+receiver+"/accion", bytes.NewBuffer(bodyJson))
	if err != nil {
		logger.Error("Error al hacer la request")
	}
	if response.StatusCode != http.StatusOK {
		logger.Error("Respuesta recibida: " + response.Status)
	}

	fmt.Print(response)
}
