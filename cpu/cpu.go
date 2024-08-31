package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func main() {
	http.HandleFunc("POST /cpu", GenerateSendResponse)
	http.HandleFunc("/", BadRequest)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	GenerateRequest("kernel")
	GenerateRequest("memoria")
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	jsonError, err := json.MarshalIndent("Ruta invalida", "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(jsonError)
}

func GenerateSendResponse(w http.ResponseWriter, r *http.Request) {
	var request BodyRequest
	requestBody, err := io.ReadAll(r.Body)
	err = json.Unmarshal(requestBody, &request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response := Response{
		Request:  request,
		Response: "Solicitud recibida de " + request.Origin,
	}
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GenerateRequest(receiver string) {
	client := &http.Client{}
	body := BodyRequest{
		Message: "Hola " + receiver,
		Origin:  "Cpu",
	}
	bodyJson, err := json.MarshalIndent(body, "", "  ")
	request, err := http.NewRequest("POST", "http://localhost:8080/"+receiver, bytes.NewBuffer(bodyJson))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	if response.StatusCode != http.StatusOK {
		return
	}

	responseBody, err := io.ReadAll(response.Body)
	fmt.Println(string(responseBody))
}
