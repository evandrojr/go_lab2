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
	go_otlp "go.opentelemetry.io/otel/exporters/zipkin"
	oteltrace "go.opentelemetry.io/otel/trace"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	otel "go.opentelemetry.io/otel"
	otelprop "go.opentelemetry.io/otel/propagation"
	otelresource "go.opentelemetry.io/otel/sdk/resource"
	otelsdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
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

var fetchCoordinates = func(ctx context.Context, cep, token string) (*CepAbertoResponse, error) {
	url := fmt.Sprintf("https://www.cepaberto.com/api/v3/cep?cep=%s", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao realizar requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resposta inesperada da API: %s", resp.Status)
	}

	// Nessa API quando o CEP não é encontrado, a resposta é 200 e o corpo é vazio.
	var data CepAbertoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		// ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return nil, fmt.Errorf("can not find zipcode")
	}

	if data.Latitude == "" || data.Longitude == "" {
		return nil, fmt.Errorf("can not find zipcode")
	}

	return &data, nil
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

func initTracer() (func(), error) {
	zipkinURL := "http://localhost:9411/api/v2/spans"
	exporter, err := go_otlp.New(zipkinURL)
	if err != nil {
		return nil, err
	}
	resource := otelresource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("servico-b"),
	)
	provider := otelsdktrace.NewTracerProvider(
		otelsdktrace.WithBatcher(exporter),
		otelsdktrace.WithResource(resource),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(otelprop.TraceContext{})
	return func() { _ = provider.Shutdown(context.Background()) }, nil
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
	router.Use(otelgin.Middleware("servico-b"))

	router.GET("/temp/:cep", func(c *gin.Context) {
		cep := c.Param("cep")

		if !validarCEP(cep) {
			c.JSON(422, gin.H{"error": "invalid zipcode."})
			return
		}

		token := os.Getenv("API_TOKEN")
		if token == "" {
			c.JSON(500, gin.H{"error": "API token not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()

		tr := otel.Tracer("servico-b")
		var span oteltrace.Span
		ctx, span = tr.Start(ctx, "buscar-coordenadas-cep")
		coordenadas, err := fetchCoordinates(ctx, cep, token)
		span.End()
		if err != nil {
			if err.Error() == "can not find zipcode" {
				c.JSON(http.StatusNotFound, gin.H{"error": "can not find zipcode"})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		_, span2 := tr.Start(ctx, "buscar-temperatura")
		temperatura, err := getTemperature(coordenadas.Latitude, coordenadas.Longitude)
		span2.End()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		temp := Temperatura{
			City:  coordenadas.Cidade.Nome,
			TempC: temperatura,
			TempF: temperatura*1.8 + 32,
			TempK: temperatura + 273,
		}
		c.JSON(http.StatusOK, temp)
	})

	router.Run(":8081")
}
