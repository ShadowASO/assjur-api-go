# Configurações de Produção
# NO ARQUIVO (.env)

# Essa variável de ambiente é lida pelo servidor na inicialização e utilizada para configurar o CORS. 
# Ficar atento para a configuração dessa varável, pois o CORS irá bloquear qualquer origem diferente
CORS_ORIGINS_ALLOWED = "http://localhost:3002,https://wiseworld.com.br"

# Configurações do servidor de api. o nginx está configurado para a porta 4001.
SERVER_HOST='localhost'
SERVER_PORT=':4001'

# Configurações para o docker compose
POSTGRES_DB=assjurdb
POSTGRES_USER=assjurpg
POSTGRES_PASSWORD=
# HOST DO servidor postgresql:
# VPS
POSTGRES_HOST='191.101.71.18'
# LOCAL
#POSTGRES_HOST='localhost'
# CONTAINER
#POSTGRES_HOST='dcs-postgres'
POSTGRES_PORT=7432

# Modo de Execução da Aplicação em Produção
GIN_MODE=release

# ATENÇÃO - Modificar a porta nos arquivos:
Dockerfile
docker-compose-server.yml
