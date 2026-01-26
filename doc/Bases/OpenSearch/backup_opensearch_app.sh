#Este escript realiza o backup do opensearch
#Instalar na VPS:
#sudo cp backup_opensearch_app.sh /usr/local/bin/backup_opensearch_app.sh
#sudo chmod 755 /usr/local/bin/backup_opensearch_app.sh
#USO: backup_opensearch_app.sh

#!/usr/bin/env bash
set -euo pipefail
#!/usr/bin/env bash
set -euo pipefail

# ============================================================
# Backup OpenSearch (APP) - snapshot dos índices essenciais
# Frequência sugerida: diário ou horário (cron/systemd timer)
#
# Índices incluídos:
#   contexto, autos, autos_temp, eventos, modelos,
#   base_doc_embedding, autos_doc_embedding, autos_json_embedding
#
# Variáveis (opcionais):
#   OS_URL="http://localhost:9200"
#   REPO="meu_backup"
#   INCLUDE_AUTOS_TEMP="true" | "false"
#   WAIT="true" | "false"
# ============================================================

OS_URL="${OS_URL:-http://localhost:9200}"
REPO="${REPO:-meu_backup}"
INCLUDE_AUTOS_TEMP="${INCLUDE_AUTOS_TEMP:-true}"
WAIT="${WAIT:-true}"

# snapshot name: app_YYYY-MM-DD_HHMMSS
SNAP="app_$(date +%F_%H%M%S)"

# Lista base de índices essenciais
INDICES="contexto,autos,eventos,modelos,base_doc_embedding,autos_doc_embedding,autos_json_embedding"
if [[ "${INCLUDE_AUTOS_TEMP}" == "true" ]]; then
  INDICES="${INDICES},autos_temp"
fi

echo "[backup] OpenSearch URL : ${OS_URL}"
echo "[backup] Repo          : ${REPO}"
echo "[backup] Snapshot      : ${SNAP}"
echo "[backup] Indices       : ${INDICES}"
echo

# --- Checagens rápidas ---
# 1) OpenSearch acessível
curl -sSf "${OS_URL}" > /dev/null

# 2) Repo existe
if ! curl -sS "${OS_URL}/_snapshot/${REPO}" | grep -q "\"${REPO}\""; then
  echo "[erro] Repositório de snapshot '${REPO}' não encontrado no OpenSearch."
  echo "       Crie o repositório antes (PUT /_snapshot/${REPO})."
  exit 1
fi

# 3) Índices essenciais existem (falha se algum estiver ausente)
missing=()
IFS=',' read -r -a arr <<< "${INDICES}"
for idx in "${arr[@]}"; do
  if ! curl -sS -o /dev/null -w "%{http_code}" "${OS_URL}/${idx}" | grep -qE '^(200|401|403)$'; then
    # 200 ok; 401/403: existe mas exige auth (caso raro em localhost)
    missing+=("${idx}")
  fi
done

if (( ${#missing[@]} > 0 )); then
  echo "[erro] Índices ausentes (não será feito snapshot): ${missing[*]}"
  exit 1
fi

# --- Snapshot ---
WAIT_QS=""
if [[ "${WAIT}" == "true" ]]; then
  WAIT_QS="wait_for_completion=true"
else
  WAIT_QS="wait_for_completion=false"
fi

resp="$(curl -sS -X PUT "${OS_URL}/_snapshot/${REPO}/${SNAP}?${WAIT_QS}" \
  -H "Content-Type: application/json" \
  -d "{
    \"indices\": \"${INDICES}\",
    \"ignore_unavailable\": false,
    \"include_global_state\": false
  }")"

# --- Resultado ---
# Se WAIT=true, vem "snapshot": {..., "state":"SUCCESS"}.
# Se WAIT=false, vem "accepted": true.
if echo "${resp}" | grep -q '"state"[[:space:]]*:[[:space:]]*"SUCCESS"'; then
  echo "[ok] Snapshot concluído com SUCCESS."
elif echo "${resp}" | grep -q '"accepted"[[:space:]]*:[[:space:]]*true'; then
  echo "[ok] Snapshot aceito (execução assíncrona no cluster)."
else
  echo "[erro] Resposta inesperada ao criar snapshot:"
  echo "${resp}"
  exit 1
fi

echo
echo "[info] Últimos snapshots no repositório '${REPO}':"
curl -sS "${OS_URL}/_cat/snapshots/${REPO}?v&s=start_epoch:desc" | head -n 12

