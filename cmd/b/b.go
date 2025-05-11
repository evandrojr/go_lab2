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
)

type CepAbertoResponse struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

type OpenMeteoResponse struct {
	CurrentWeather struct {
		Temperature float64 `json:"temperature"`
	} `json:"current_weather"`
}

type Temperatura struct {
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

func main() {
	// Carrega variáveis do .env
	_ = godotenv.Load()

	token := os.Getenv("API_TOKEN")
	fmt.Println("[DEBUG] API_TOKEN carregado:", token)

	router := gin.Default()

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

		coordenadas, err := fetchCoordinates(ctx, cep, token)
		if err != nil {
			if err.Error() == "can not find zipcode" {
				c.JSON(http.StatusNotFound, gin.H{"error": "can not find zipcode"})
				return
			}
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		temperatura, err := getTemperature(coordenadas.Latitude, coordenadas.Longitude)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		temp := Temperatura{
			TempC: temperatura,
			TempF: temperatura*1.8 + 32,
			TempK: temperatura + 273,
		}
		c.JSON(http.StatusOK, temp)
	})

	router.Run(":8081")
}
