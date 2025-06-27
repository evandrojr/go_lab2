package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	otelhttp "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	otel "go.opentelemetry.io/otel"
	go_otlp "go.opentelemetry.io/otel/exporters/zipkin"
	otelprop "go.opentelemetry.io/otel/propagation"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type CepAbertoResponse struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Cidade    struct {
		Nome string `json:"nome"`
	} `json:"cidade"`
}

type OpenMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

type Temperatura struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func validarCEP(cep string) bool {
	if len(cep) != 8 {
		return false
	}
	for _, c := range cep {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func initTracer() (func(), error) {
	zipkinURL := "http://zipkin:9411/api/v2/spans"
	exporter, err := go_otlp.New(zipkinURL)
	if err != nil {
		return nil, err
	}
	resource := otelresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("servico-a"),
	)
	provider := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter),
		otelsdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(otelprop.TraceContext{})
	return func() { _ = provider.Shutdown(context.Background()) }, nil
}

// Instrumenta fetchTemp para propagação do contexto OTEL
var fetchTemp = func(ctx context.Context, cep, token string) (*Temperatura, error) {
	url := fmt.Sprintf("http://servico-b:8081/temp/%s", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta inesperada da API: %s", resp.Status)
	}

	// Faça o parse da resposta em uma estrutura de Temperatura
	var temp Temperatura
	if err := json.NewDecoder(resp.Body).Decode(&temp); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}
	return &temp, nil

	// return &Temperatura{}, nil
}

var getTemperature = func(latStr, lonStr string) (float64, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, fmt.Errorf("latitude inválida: %w", err)
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, fmt.Errorf("longitude inválida: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current_weather=true", lat, lon)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("erro ao realizar requisição: %w", err)
	}
	defer resp.Body.Close()

	var data OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return data.CurrentWeather.Temperature, nil
}

func main() {
	// Carrega variáveis do .env
	_ = godotenv.Load()

	shutdown, err := initTracer()
	if err != nil {
		panic(err)
	}
	defer shutdown()

	token := os.Getenv("API_TOKEN")
	fmt.Println("[DEBUG] API_TOKEN carregado:", token)

	router := gin.Default()
	router.Use(otelgin.Middleware("servico-a"))

	router.POST("/temp/:cep", func(c *gin.Context) {
		cep := c.Param("cep")

		if !validarCEP(cep) {
			c.JSON(422, gin.H{"error": "invalid zipcode."})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()

		tr := otel.Tracer("servico-a")
		var span oteltrace.Span
		ctx, span = tr.Start(ctx, "chamada-servico-b")
		temp, err := fetchTemp(ctx, cep, token)
		span.End()
		if err != nil {
			if err.Error() == "can not find zipcode" {
				c.JSON(http.StatusNotFound, gin.H{"error": "can not find zipcode"})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, *temp)
	})

	router.Run(":8080")
}
