package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

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

var tracer trace.Tracer

func initProvider() {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("servico_b"),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("servico_b")
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "weatherHandler")
	defer span.End()

	cep := r.URL.Query().Get("cep")

	res, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		http.Error(w, `{"message": "can not find zip code"}`, http.StatusNotFound)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
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

	res, err = http.Get(fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=602ac96551be4db2b0112256243006&q=%s&aqi=no", location))
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

func main() {
	initProvider()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/weather", weatherHandler)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", r))
}
