#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Backup OpenSearch - APP ou FULL
#
# Modos:
#   APP  -> snapshot dos índices essenciais da aplicação
#   FULL -> snapshot completo do OpenSearch
#
# Uso:
#   ./backup_opensearch.sh
#   ./backup_opensearch.sh --app
#   ./backup_opensearch.sh --full
#
# Variáveis opcionais:
#   OS_URL="http://localhost:9200"
#   REPO="meu_backup"
#   BACKUP_MODE="app" | "full"
#   INCLUDE_AUTOS_TEMP="true" | "false"
#   WAIT="true" | "false"
#   INCLUDE_GLOBAL_STATE="true" | "false"
#
# Exemplos:
#   BACKUP_MODE=full ./backup_opensearch.sh
#   INCLUDE_AUTOS_TEMP=false ./backup_opensearch.sh --app
#   WAIT=false ./backup_opensearch.sh --full
# ============================================================

OS_URL="${OS_URL:-http://localhost:9200}"
REPO="${REPO:-meu_backup}"
BACKUP_MODE="${BACKUP_MODE:-app}"
INCLUDE_AUTOS_TEMP="${INCLUDE_AUTOS_TEMP:-true}"
WAIT="${WAIT:-true}"

# No modo FULL, normalmente faz sentido incluir o estado global.
# No modo APP, normalmente NÃO se inclui o estado global.
INCLUDE_GLOBAL_STATE="${INCLUDE_GLOBAL_STATE:-}"

# ------------------------------------------------------------
# Tratamento de argumentos
# ------------------------------------------------------------
usage() {
  cat <<EOF
Uso:
  $0 [--app | --full]

Opções:
  --app       Faz backup apenas dos índices essenciais da aplicação.
  --full      Faz backup completo do OpenSearch.
  -h, --help  Exibe esta ajuda.

Variáveis:
  OS_URL=http://localhost:9200
  REPO=meu_backup
  BACKUP_MODE=app|full
  INCLUDE_AUTOS_TEMP=true|false
  WAIT=true|false
  INCLUDE_GLOBAL_STATE=true|false
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --app)
      BACKUP_MODE="app"
      shift
      ;;
    --full)
      BACKUP_MODE="full"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[erro] Argumento desconhecido: $1"
      echo
      usage
      exit 1
      ;;
  esac
done

BACKUP_MODE="$(echo "${BACKUP_MODE}" | tr '[:upper:]' '[:lower:]')"

if [[ "${BACKUP_MODE}" != "app" && "${BACKUP_MODE}" != "full" ]]; then
  echo "[erro] BACKUP_MODE inválido: ${BACKUP_MODE}"
  echo "       Use: app ou full"
  exit 1
fi

# ------------------------------------------------------------
# Definição do snapshot
# ------------------------------------------------------------
TIMESTAMP="$(date +%F_%H%M%S)"

if [[ "${BACKUP_MODE}" == "full" ]]; then
  SNAP="full_${TIMESTAMP}"
  INDICES="*"

  if [[ -z "${INCLUDE_GLOBAL_STATE}" ]]; then
    INCLUDE_GLOBAL_STATE="true"
  fi
else
  SNAP="app_${TIMESTAMP}"

  INDICES="contexto,autos,eventos,modelos,base_doc_embedding,autos_doc_embedding,autos_json_embedding"

  if [[ "${INCLUDE_AUTOS_TEMP}" == "true" ]]; then
    INDICES="${INDICES},autos_temp"
  fi

  if [[ -z "${INCLUDE_GLOBAL_STATE}" ]]; then
    INCLUDE_GLOBAL_STATE="false"
  fi
fi

# ------------------------------------------------------------
# Validação das flags booleanas
# ------------------------------------------------------------
validate_bool() {
  local name="$1"
  local value="$2"

  if [[ "${value}" != "true" && "${value}" != "false" ]]; then
    echo "[erro] ${name} deve ser true ou false. Valor recebido: ${value}"
    exit 1
  fi
}

validate_bool "INCLUDE_AUTOS_TEMP" "${INCLUDE_AUTOS_TEMP}"
validate_bool "WAIT" "${WAIT}"
validate_bool "INCLUDE_GLOBAL_STATE" "${INCLUDE_GLOBAL_STATE}"

echo "[backup] OpenSearch URL       : ${OS_URL}"
echo "[backup] Repo                : ${REPO}"
echo "[backup] Modo                : ${BACKUP_MODE}"
echo "[backup] Snapshot            : ${SNAP}"
echo "[backup] Índices             : ${INDICES}"
echo "[backup] Include global state: ${INCLUDE_GLOBAL_STATE}"
echo "[backup] Wait completion     : ${WAIT}"
echo

# ------------------------------------------------------------
# Checagens rápidas
# ------------------------------------------------------------

# 1) OpenSearch acessível
if ! curl -sSf "${OS_URL}" > /dev/null; then
  echo "[erro] Não foi possível acessar o OpenSearch em: ${OS_URL}"
  exit 1
fi

# 2) Repositório existe
if ! curl -sS "${OS_URL}/_snapshot/${REPO}" | grep -q "\"${REPO}\""; then
  echo "[erro] Repositório de snapshot '${REPO}' não encontrado no OpenSearch."
  echo "       Crie o repositório antes, por exemplo:"
  echo
  echo "       curl -X PUT '${OS_URL}/_snapshot/${REPO}' \\"
  echo "         -H 'Content-Type: application/json' \\"
  echo "         -d '{"
  echo "           \"type\": \"fs\","
  echo "           \"settings\": {"
  echo "             \"location\": \"/backup\""
  echo "           }"
  echo "         }'"
  echo
  exit 1
fi

# 3) Verificação dos índices apenas no modo APP
# No modo FULL usamos "*", então não faz sentido testar índice por índice.
if [[ "${BACKUP_MODE}" == "app" ]]; then
  missing=()

  IFS=',' read -r -a arr <<< "${INDICES}"

  for idx in "${arr[@]}"; do
    http_code="$(curl -sS -o /dev/null -w "%{http_code}" "${OS_URL}/${idx}")"

    if [[ ! "${http_code}" =~ ^(200|401|403)$ ]]; then
      missing+=("${idx}")
    fi
  done

  if (( ${#missing[@]} > 0 )); then
    echo "[erro] Índices ausentes. O snapshot APP não será feito."
    echo "       Ausentes: ${missing[*]}"
    echo
    echo "       Se você quiser ignorar índices ausentes, ajuste o script para usar:"
    echo "       \"ignore_unavailable\": true"
    exit 1
  fi
fi

# ------------------------------------------------------------
# Montagem da query string
# ------------------------------------------------------------
if [[ "${WAIT}" == "true" ]]; then
  WAIT_QS="wait_for_completion=true"
else
  WAIT_QS="wait_for_completion=false"
fi

# ------------------------------------------------------------
# Snapshot
# ------------------------------------------------------------
echo "[backup] Iniciando snapshot..."

resp="$(curl -sS -X PUT "${OS_URL}/_snapshot/${REPO}/${SNAP}?${WAIT_QS}" \
  -H "Content-Type: application/json" \
  -d "{
    \"indices\": \"${INDICES}\",
    \"ignore_unavailable\": false,
    \"include_global_state\": ${INCLUDE_GLOBAL_STATE}
  }")"

# ------------------------------------------------------------
# Resultado
# ------------------------------------------------------------
if echo "${resp}" | grep -q '"state"[[:space:]]*:[[:space:]]*"SUCCESS"'; then
  echo "[ok] Snapshot concluído com SUCCESS."
elif echo "${resp}" | grep -q '"accepted"[[:space:]]*:[[:space:]]*true'; then
  echo "[ok] Snapshot aceito. Execução assíncrona no cluster."
else
  echo "[erro] Resposta inesperada ao criar snapshot:"
  echo "${resp}"
  exit 1
fi

echo
echo "[info] Últimos snapshots no repositório '${REPO}':"
curl -sS "${OS_URL}/_cat/snapshots/${REPO}?v&s=start_epoch:desc" | head -n 12
