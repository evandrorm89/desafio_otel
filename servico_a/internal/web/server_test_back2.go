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
	reqMap := map[string]string{
		"cep": "07096240",
	}
	reqBody, err := json.Marshal(reqMap)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc((*Webserver).HandleRequest)

	handler.ServeHttp(rr, req)

	var resp struct {
		Cidade string  `json:"city"`
		TempC  float64 `json:"temp_C"`
		TempF  float64 `json:"temp_F"`
		TempK  float64 `json:"temp_K"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, resp)
}
