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
		w.Write([]byte(`{"city": "São Paulo", "Temp_c": 27.8, "Temp_f: 82, "Temp_k: 300.8}`))
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
	// resp, err := http.Post("/", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Println("Error making POST request:", err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var resp struct {
		Cidade string  `json:"city"`
		TempC  float64 `json:"temp_C"`
		TempF  float64 `json:"temp_F"`
		TempK  float64 `json:"temp_K"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(resp)
}
