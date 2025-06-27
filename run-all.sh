#!/bin/bash

# Remove containers antigos, se existirem
docker rm -f zipkin servico-a servico-b 2>/dev/null

# Subir Zipkin
docker run -d --name zipkin -p 9411:9411 openzipkin/zipkin

# Build Serviço B
cd cmd/b
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o servico-b .
cd ../..

# Build Serviço A
cd cmd/a
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o servico-a .
cd ../..

# Subir Serviço B (modo interativo para debug, mostra logs imediatamente)
echo "--- LOGS servico-b ---"
docker run --rm \
  --name servico-b \
  --env API_TOKEN=a9ff0b35dd43008c20bbc78465042df9 \
  -p 8081:8081 \
  -v "$(pwd)/cmd/b/servico-b:/servico-b" \
  --network host \
  alpine:latest sh -c "chmod +x /servico-b && /servico-b" &

# Subir Serviço A (modo interativo para debug, mostra logs imediatamente)
echo "--- LOGS servico-a ---"
docker run --rm \
  --name servico-a \
  --env API_TOKEN=a9ff0b35dd43008c20bbc78465042df9 \
  -p 8080:8080 \
  -v "$(pwd)/cmd/a/servico-a:/servico-a" \
  --network host \
  alpine:latest sh -c "chmod +x /servico-a && /servico-a" &

wait