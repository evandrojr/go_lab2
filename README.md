# Projeto Clima - Documentação de Ambiente de Desenvolvimento

## Pré-requisitos
- Docker e Docker Compose instalados
- Go 1.20+ instalado (opcional, apenas para builds locais)

## Configuração de variáveis de ambiente

1. Copie o arquivo `.env.example` para `.env` na raiz do projeto:
   ```sh
   cp .env.example .env
   ```
2. Edite o arquivo `.env` e configure o valor da variável `API_TOKEN` conforme necessário. Mas já deixei um que funciona.

## Subindo o ambiente com Docker Compose

1. **Na raiz do projeto, execute:**
   ```sh
   docker compose up --build
   ```
   Isso irá:
   - Subir o Zipkin em http://localhost:9411
   - Subir o Serviço B em http://localhost:8081
   - Subir o Serviço A em http://localhost:8080

2. **Para parar e remover os containers:**
   ```sh
   ./stop-all.sh
   ```


## Testando a API

- Use o arquivo `clima.http` (VSCode REST Client) ou ferramentas como curl/Postman:

  - **Exemplo de requisição para Serviço A:**
    ```sh
    curl -X POST http://localhost:8080/temp/41830460
    ```
  - **Exemplo de requisição para Serviço B:**
    ```sh
    curl http://localhost:8081/temp/41830460
    ```

## Observabilidade (Tracing)

- Acesse o Zipkin em [http://localhost:9411](http://localhost:9411)
- Faça requisições para os serviços e clique em "Run Query" para visualizar os traces e spans distribuídos.

---

**Dúvidas ou problemas?**
Consulte os logs dos containers para debug.
