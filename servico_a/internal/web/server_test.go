package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestHandler(t *testing.T) {
	serverMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"city": "São Paulo", "temp_C": 27.8, "temp_F": 82, "temp_K": 300.8}`))
	}))
	defer serverMock.Close()

	tracer := otel.Tracer("microservice-tracer-mock")

	serviceData := &ServiceData{
		ExternalCallURL: serverMock.URL,
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	server := NewServer(serviceData)
	router := server.CreateServer()

	reqMap := map[string]string{
		"cep": "07096240",
	}
	reqBody, _ := json.Marshal(reqMap)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error making POST request:", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp struct {
		City   string  `json:"city"`
		Temp_c float64 `json:"temp_C"`
		Temp_f float64 `json:"temp_F"`
		Temp_k float64 `json:"temp_K"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, resp)
	assert.Equal(t, "São Paulo", resp.City)
	assert.Equal(t, 27.8, resp.Temp_c)
	assert.Equal(t, 82.0, resp.Temp_f)
	assert.Equal(t, 300.8, resp.Temp_k)
}

func TestHandlerZipCodeNotFound(t *testing.T) {
	serverMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "can not find zip code"}`))
	}))
	defer serverMock.Close()

	tracer := otel.Tracer("microservice-tracer-mock")

	serviceData := &ServiceData{
		ExternalCallURL: serverMock.URL,
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	server := NewServer(serviceData)
	router := server.CreateServer()

	reqMap := map[string]string{
		"cep": "00000000",
	}
	reqBody, _ := json.Marshal(reqMap)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
	// resp, err := http.Post("/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error making POST request:", err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp struct {
		City   string  `json:"city"`
		Temp_c float64 `json:"temp_C"`
		Temp_f float64 `json:"temp_F"`
		Temp_k float64 `json:"temp_K"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, resp)
}

func TestHandlerInvalidZipCode(t *testing.T) {
	serverMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "invalid zipcode"}`))
	}))
	defer serverMock.Close()

	tracer := otel.Tracer("microservice-tracer-mock")

	serviceData := &ServiceData{
		ExternalCallURL: serverMock.URL,
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	server := NewServer(serviceData)
	router := server.CreateServer()

	reqMap := map[string]string{
		"cep": "abvc",
	}
	reqBody, _ := json.Marshal(reqMap)
	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error making POST request:", err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp struct {
		City   string  `json:"city"`
		Temp_c float64 `json:"temp_C"`
		Temp_f float64 `json:"temp_F"`
		Temp_k float64 `json:"temp_K"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Empty(t, resp)
}
