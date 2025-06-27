# Etapa de construção
FROM golang:1.24.2-alpine AS builder

# Definir o diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod tidy

# Copiar o código-fonte
COPY . .

# Build para Serviço B
RUN CGO_ENABLED=0 GOOS=linux go build -o servico-b ./cmd/b
# Build para Serviço A
RUN CGO_ENABLED=0 GOOS=linux go build -o servico-a ./cmd/a

# Etapa de execução
FROM gcr.io/distroless/base-debian11

# Definir o diretório de trabalho
WORKDIR /root/

# Definir variável de ambiente
ENV API_TOKEN=a9ff0b35dd43008c20bbc78465042df9

# Copia ambos os binários
COPY --from=builder /app/servico-b ./servico-b
COPY --from=builder /app/servico-a ./servico-a

# Expor a porta que o aplicativo irá escutar
EXPOSE 8080 8081

# O entrypoint será definido no docker-compose.yml
CMD ["/bin/sh"]
