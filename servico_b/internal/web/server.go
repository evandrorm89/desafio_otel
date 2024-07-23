package web

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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

type ViaCepResponse struct {
	Localidade string `json:localidade`
}

type WeatherReport struct {
	Current WeatherResponse `json:current`
}

type Current struct {
	Temp_c float64 `json:temp_c`
	Temp_f float64 `json:temp_f`
}

type WeatherResponse struct {
	Temp_c float64 `json:temp_C`
	Temp_f float64 `json:temp_F`
	Temp_k float64 `json:temp_K`
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
	CepURL             string
	WeatherURL         string
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
	if err != nil {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}
	cep := request["cep"]

	// Buscar cep na api da viacep
	var req *http.Request
	endpoint := fmt.Sprintf(h.ServiceData.CepURL, cep)
	req, err = http.NewRequestWithContext(ctx, h.ServiceData.ExternalCallMethod, endpoint, nil)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "can not find zip code", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var c ViaCepResponse
	err = json.Unmarshal(body, &c)
	if err != nil {
		http.Error(w, `{"message": "Erro interno"}`, http.StatusInternalServerError)
		return
	}

	if c.Localidade == "" {
		http.Error(w, `{"message": "can not find zip code"}`, http.StatusNotFound)
		return
	}

	location := url.QueryEscape(c.Localidade)

	// res, err := http.Get(fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=602ac96551be4db2b0112256243006&q=%s&aqi=no", location))
	res, err := http.Get(fmt.Sprintf(h.ServiceData.WeatherURL, location))
	if err != nil {
		http.Error(w, `{"message": "Erro ao achar o tempo atual para a localidade informada"}`, http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)

	if err != nil {
		http.Error(w, `{"message": "Erro ao achar o tempo atual para a localidade informada"}`, http.StatusInternalServerError)
		return
	}

	var t WeatherReport
	err = json.Unmarshal(body, &t)
	if err != nil {
		http.Error(w, `{"message": "Erro ao achar o tempo atual para a localidade informada"}`, http.StatusInternalServerError)
		fmt.Println(err)
		return
	}

	tempC := t.Current.Temp_c
	tempF := t.Current.Temp_f
	tempK := t.Current.Temp_c + 273.0

	response := WeatherResponse{
		Temp_c: tempC,
		Temp_f: tempF,
		Temp_k: tempK,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
