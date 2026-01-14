#!/bin/bash

# Configurações
CONTAINER=dcs-postgres
BACKUP_FILE=${1:-backup.sql}  # Pode passar o nome do backup como argumento, ex: ./restore.sh meu_backup.sql

# Checa se o container está rodando
echo "Verificando se o container $CONTAINER está rodando..."
if ! docker ps | grep -q $CONTAINER; then
  echo "Container $CONTAINER não está rodando. Iniciando..."
  docker compose -f docker-compose-postgresql.yml up -d
  sleep 5  # Aguarda alguns segundos para o banco inicializar
fi

# Copia o arquivo de backup para dentro do container
echo "Copiando backup $BACKUP_FILE para o container..."
docker cp "$BACKUP_FILE" $CONTAINER:/tmp/backup.sql

# Executa o restore
echo "Restaurando backup..."
docker exec -e PGPASSWORD=${POSTGRES_PASSWORD} $CONTAINER \
  psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/backup.sql

# Limpa o arquivo dentro do container
docker exec $CONTAINER rm /tmp/backup.sql

echo "Restauração concluída!"
