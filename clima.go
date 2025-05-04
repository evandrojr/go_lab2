package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Estrutura para a resposta da API do CEP Aberto
type CepAbertoResponse struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Função para buscar coordenadas a partir do CEP
func fetchCoordinates(ctx context.Context, cep, token string) (*CepAbertoResponse, error) {
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

	var data CepAbertoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %w", err)
	}

	return &data, nil
}

func main() {
	router := gin.Default()

	router.GET("/coordenadas/:cep", func(c *gin.Context) {
		cep := c.Param("cep")

		// Validação simples do CEP
		if len(cep) != 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "CEP inválido. Use exatamente 8 dígitos numéricos."})
			return
		}

		// Token de acesso à API do CEP Aberto
		token := "a9ff0b35dd43008c20bbc78465042df9" // Substitua pelo seu token

		// Contexto com timeout de 3 segundos
		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()

		coordenadas, err := fetchCoordinates(ctx, cep, token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, coordenadas)
	})

	router.Run(":8080")
}
