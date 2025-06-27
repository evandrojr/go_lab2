#!/bin/bash

echo "Parando e removendo containers servico-a, servico-b e zipkin..."
docker rm -f servico-a servico-b zipkin 2>/dev/null

echo "Pronto. Todos os containers foram parados e removidos."
