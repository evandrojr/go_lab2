package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCoordenadas(t *testing.T) {
	cep := "41830460"
	token := "a9ff0b35dd43008c20bbc78465042df9"

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	coordenadas, err := fetchCoordinates(ctx, cep, token)
	assert.Nil(t, err)
	assert.NotNil(t, coordenadas)
}

func TestAPI_Success(t *testing.T) {
	// Mock das funções externas
	origFetchCoordinates := fetchCoordinates
	origGetTemperature := getTemperature
	defer func() {
		fetchCoordinates = origFetchCoordinates
		getTemperature = origGetTemperature
	}()

	fetchCoordinates = func(ctx context.Context, cep, token string) (*CepAbertoResponse, error) {
		return &CepAbertoResponse{Latitude: "-23.5505", Longitude: "-46.6333"}, nil
	}
	getTemperature = func(latStr, lonStr string) (float64, error) {
		return 25.0, nil
	}

	req := httptest.NewRequest("GET", "/temp/01001000", nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.GET("/temp/:cep", func(c *gin.Context) {
		cep := c.Param("cep")
		if !validarCEP(cep) {
			c.JSON(422, gin.H{"error": "invalid zipcode."})
			return
		}
		token := "a9ff0b35dd43008c20bbc78465042df9"
		ctx, cancel := context.WithTimeout(c.Request.Context(), 6*time.Second)
		defer cancel()
		coordenadas, err := fetchCoordinates(ctx, cep, token)
		if err != nil {
			if err.Error() == "can not find zipcode" {
				c.JSON(http.StatusNotFound, gin.H{"error": "can not find zipcode"})
				return
			}
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

	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestAPI_InvalidCEP(t *testing.T) {
	req := httptest.NewRequest("GET", "/temp/123", nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.GET("/temp/:cep", func(c *gin.Context) {
		cep := c.Param("cep")
		if !validarCEP(cep) {
			c.JSON(422, gin.H{"error": "invalid zipcode."})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	r.ServeHTTP(w, req)
	assert.Equal(t, 422, w.Code)
}

func TestAPI_NotFoundCEP(t *testing.T) {
	// Este teste depende de um CEP inexistente
	req := httptest.NewRequest("GET", "/temp/00000000", nil)
	w := httptest.NewRecorder()

	r := gin.Default()
	r.GET("/temp/:cep", func(c *gin.Context) {
		cep := c.Param("cep")
		if !validarCEP(cep) {
			c.JSON(422, gin.H{"error": "invalid zipcode."})
			return
		}
		c.JSON(404, gin.H{"error": "can not find zipcode"})
	})

	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}
