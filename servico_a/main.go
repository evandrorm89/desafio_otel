package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

func initProvider() {
	// Configurar o exportador Zipkin
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		log.Fatal(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("servico_a"),
		)),
	)

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("servico_a")
}

func cepHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "cepHandler")
	defer span.End()

	var request map[string]string
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || len(request["cep"]) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Encaminhar para o Serviço B
	client := &http.Client{}
	reqBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "http://servico_b:8080/weather", bytes.NewBuffer(reqBody))
	req = req.WithContext(ctx)
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func main() {
	initProvider()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Post("/cep", cepHandler)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8081", r))
}
