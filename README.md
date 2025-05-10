# Clima por CEP

Este projeto é um serviço em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em Celsius, Fahrenheit e Kelvin).

## Pré-requisitos
- Go 1.18+
- Docker (opcional, para rodar via container)

## Configuração
1. Copie o arquivo `.env.example` para `.env` na raiz do projeto:
   ```sh
   cp .env.example .env
   ```
2. Edite o arquivo `.env` e coloque o seu token da API CepAberto:
   ```env
   API_TOKEN=a9ff0b35dd43008c20bbc78465042df9
   ```

## Executando localmente

### Usando Go
```sh
go run clima.go
```
O serviço estará disponível em http://localhost:8080

### Usando Docker
```sh
docker build -t clima .
docker run --env-file .env -p 8080:8080 clima
```

## Como usar
Faça uma requisição GET para:
```
GET /temp/{cep}
```
Exemplo:
```
curl http://localhost:8080/temp/01001000
```

## Respostas
- **200**: Sucesso
  ```json
  { "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.1 }
  ```
- **422**: CEP inválido
  ```json
  { "error": "invalid zipcode." }
  ```
- **404**: CEP não encontrado
  ```json
  { "error": "can not find zipcode" }
  ```
- **500**: Token da API não configurado
  ```json
  { "error": "API token not configured" }
  ```

## Testes
Para rodar os testes:
```sh
go test -v
```
