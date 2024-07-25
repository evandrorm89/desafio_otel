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
	tracer := otel.Tracer("microservice-tracer-mock")

	serviceData := &ServiceData{
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
		CepURL:          "http://viacep.com.br/ws/%s/json/",
		WeatherURL:      "http://api.weatherapi.com/v1/current.json?key=602ac96551be4db2b0112256243006&q=%s&aqi=no",
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
	assert.Equal(t, "Guarulhos", resp.City)
}

func TestHandlerZipCodeNotFound(t *testing.T) {
	tracer := otel.Tracer("microservice-tracer-mock")

	serviceData := &ServiceData{
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
		CepURL:          "http://viacep.com.br/ws/%s/json/",
		WeatherURL:      "http://api.weatherapi.com/v1/current.json?key=602ac96551be4db2b0112256243006&q=%s&aqi=no",
	}

	server := NewServer(serviceData)
	router := server.CreateServer()

	reqMap := map[string]string{
		"cep": "00000000",
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

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, resp)
}
