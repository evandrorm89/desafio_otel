package web

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type Webserver struct {
	ServiceData *ServiceData
}

// NewServer creates a new server instance
func NewServer(serviceData *ServiceData) *Webserver {
	return &Webserver{
		ServiceData: serviceData,
	}
}

// createServer creates a new server instance with go chi router
func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	// promhttp
	// router.Handle("/metrics", promhttp.Handler())
	router.Post("/", we.HandleRequest)
	return router
}

type ServiceData struct {
	Title              string
	ExternalCallMethod string
	ExternalCallURL    string
	RequestNameOTEL    string
	OTELTracer         trace.Tracer
}

func (h *Webserver) HandleRequest(w http.ResponseWriter, r *http.Request) {
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	ctx, span := h.ServiceData.OTELTracer.Start(ctx, "Chamada externa"+h.ServiceData.RequestNameOTEL)
	defer span.End()

	var request map[string]string
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || len(request["cep"]) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Encaminhar para o Serviço B
	var req *http.Request
	reqBody, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "error in the body", http.StatusInternalServerError)
		return
	}
	req, err = http.NewRequestWithContext(ctx, h.ServiceData.ExternalCallMethod, h.ServiceData.ExternalCallURL, bytes.NewBuffer(reqBody))
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

}
