#USO: ./backup_opensearch_app.sh

cat > backup_opensearch_app.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

OS_URL="${OS_URL:-http://localhost:9200}"
REPO="${REPO:-meu_backup}"
SNAP="app_$(date +%F_%H%M)"

curl -sS -X PUT "$OS_URL/_snapshot/$REPO/$SNAP?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }' | cat

echo
curl -sS "$OS_URL/_cat/snapshots/$REPO?v"
EOF

chmod +x backup_opensearch_app.sh

