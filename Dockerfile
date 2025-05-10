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

# Compilar o aplicativo
RUN CGO_ENABLED=0 GOOS=linux go build -o go_lab1 .

# Etapa de execução
FROM gcr.io/distroless/base-debian11

# Definir o diretório de trabalho
WORKDIR /root/

# Definir variável de ambiente
ENV API_TOKEN=a9ff0b35dd43008c20bbc78465042df9

# Copiar o binário compilado
COPY --from=builder /app/go_lab1 .

# Expor a porta que o aplicativo irá escutar
EXPOSE 8080

# Comando para iniciar o aplicativo
CMD ["/root/go_lab1"]
