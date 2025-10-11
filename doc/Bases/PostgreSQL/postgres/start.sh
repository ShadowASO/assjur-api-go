#!/bin/bash

# Nome da rede e do volume
NETWORK_NAME="local-network-ia"
VOLUME_NAME="vpgsql"

# Verifica se a rede já existe
if ! docker network ls | grep -q "$NETWORK_NAME"; then
    echo "Criando a rede $NETWORK_NAME..."
    docker network create --driver bridge "$NETWORK_NAME"
else
    echo "A rede $NETWORK_NAME já existe."
fi

# Verifica se o volume já existe
if ! docker volume ls | grep -q "$VOLUME_NAME"; then
    echo "Criando o volume $VOLUME_NAME..."
    docker volume create "$VOLUME_NAME"
else
    echo "O volume $VOLUME_NAME já existe."
fi

# Executa o docker-compose
docker compose -f docker-compose-postgresql.yml up -d