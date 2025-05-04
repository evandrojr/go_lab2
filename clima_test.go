package main

import (
	"context"
	"testing"
	"time"

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
