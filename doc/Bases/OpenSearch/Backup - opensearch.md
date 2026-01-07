history | grep backup
 1186  cd assjur-api-go_backup_20250511195447/
 1275  ls *_backup_*
 1276  rm -rf *_backup_*
 1365  mkdir -p /home/aldenor/opensearch-backup
 1366  chmod 777 /home/aldenor/opensearch-backup
 1367  docker run --rm -v /home/aldenor/opensearch-backup:/backup alpine touch /backup/teste.txt
 1368  sudo docker run --rm -v /home/aldenor/opensearch-backup:/backup alpine touch /backup/teste.txt
 1370  cd opensearch-backup/
 1394  cd /home/aldenor/opensearch-backup/
 1397  tar czf opensearch-backup.tar.gz
 1402  tar xzf opensearch-backup.tar.gz 
 1408  curl -XPUT "http://localhost:9200/_snapshot/meu_backup" -H 'Content-Type: application/json' -d'
 1411    "settings": { "location": "/backup" }
 1415  curl -XPUT "http://localhost:9200/_snapshot/meu_backup" -H 'Content-Type: application/json' -d'
 1418    "settings": { "location": "/backup" }
 1420  sudo chmod 777 /home/aldenor/opensearch-backup
 1421  curl -XPUT "http://localhost:9200/_snapshot/meu_backup" -H 'Content-Type: application/json' -d'
 1424    "settings": { "location": "/backup" }
 1426  curl -XGET "http://localhost:9200/_snapshot/meu_backup/_all"
 1427  curl -XPOST "http://localhost:9200/_snapshot/meu_backup/snapshot_1/_restore?wait_for_completion=true"
 1429  curl -XPOST "http://localhost:9200/_snapshot/meu_backup/snapshot_1/_restore?wait_for_completion=true"
 2001  ls ./opensearch-backup/
 2003  ls ./opensearch-backup/teste.txt 
 2004  cat ./opensearch-backup/teste.txt 
 2005  ls ./opensearch-backup/teste.txt 
 2006  nano ./opensearch-backup/teste.txt 
 2009  ls /var/backups/
 2023  history | grep backup
 
curl -X DELETE "http://localhost:9200/autos_temp"
 1957  curl -X PUT "http://localhost:9200/autos_temp"   -H "Content-Type: application/json"   -d '{


# Deu certo

curl -XGET "http://localhost:9200/_snapshot/meu_backup/_all"

{"snapshots":[{"snapshot":"snapshot_1","uuid":"GfebEJSMSfeeHhQcgTfWoQ","version_id":136407927,"version":"2.19.1","remote_store_index_shallow_copy":false,"indices":["autos_doc_embedding",".kibana_1","top_queries-2025.07.25-40861",".plugins-ml-config","autos_temp","top_queries-2025.07.22-40858",".opensearch-observability","autos_json_embedding","top_queries-2025.07.27-40863","autos","top_queries-2025.07.24-40860",".opensearch-sap-log-types-config","decisoes","top_queries-2025.07.23-40859","modelos","modelos_semantico",".ql-datasources","top_queries-2025.07.26-40862"],"data_streams":[],"include_global_state":true,"state":"SUCCESS","start_time":"2025-07-28T00:04:15.894Z","start_time_in_millis":1753661055894,"end_time":"2025-07-28T00:04:16.294Z","end_time_in_millis":1753661056294,"duration_in_millis":400,"failures":[],"shards":{"total":32,"failed":0,"successful":32}}]}

curl -XPUT "http://localhost:9200/_snapshot/meu_backup" -H 'Content-Type: application/json' -d' "settings": { "location": "/backup" }

PUT /_snapshot/meu_repositorio/snapshot_2026_01_06
{
  "indices": "*",
  "ignore_unavailable": true,
  "include_global_state": true
}

## CORRETO
nano /srv/assjur/opensearch/docker-compose.yml

sudo mkdir -p /home/aldenor/opensearch-backup
sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
sudo chmod -R 750 /home/aldenor/opensearch-backup

Ver se o diretório existe dentro do container:
docker exec -it os-node1 sh -lc 'ls -ld /backup && id'
docker exec -it os-node2 sh -lc 'ls -ld /backup && id'

Confirmar que o OpenSearch “enxerga” o path.repo:
curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n '"path"' -n

Se esse _verify falhar, quase sempre é permissão do diretório montado ou path.repo não aplicado.
curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"

criar um snapshot de teste (via curl): Snapshot de todos os índices
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/backup_teste_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "*",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo

Confirmar que foi criado
curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

Ver status (se rodar sem wait_for_completion)
curl -sS "http://localhost:9200/_snapshot/_status?pretty"


Uma melhoria no seu comando de inspeção (pra ver o path.repo claramente)
O grep que você usou só achou "path", mas não mostra o repo. Use:
curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n "path.repo" -n

ou (melhor, sem grep frágil):
curl -sS "http://localhost:9200/_nodes/settings?pretty" | sed -n '1,260p'


### Agora rode o snapshot de novo
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/app_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo

### E confira:
curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

O problema no seu comando

Você digitou:

curl -sS "http://localhost:9200/_snapshot/meu_backup?pretty


Perceba que:

❌ faltou fechar as aspas (") no final da URL.

Quando o shell vê uma aspa aberta, ele espera a continuação da string, por isso mostra:

>


Isso não é um prompt do curl, é o prompt do shell (bash/zsh) dizendo:

“Você abriu uma string e não terminou.”

✅ Como corrigir

Basta fechar as aspas (ou cancelar e digitar de novo).

Forma correta:
curl -sS "http://localhost:9200/_snapshot/meu_backup?pretty"


ou até sem aspas (não são obrigatórias nesse caso):

curl -sS http://localhost:9200/_snapshot/meu_backup?pretty

🛑 Como sair do prompt > sem executar nada

Se você cair nesse prompt por engano:

Ctrl + C


Isso cancela o comando e volta ao prompt normal.

🧠 Resumo rápido
Situação	Significado
> aparece	Comando incompleto
Causa comum	Aspa ", ' ou { não fechados
Solução	Fechar o que abriu ou Ctrl+C
📌 Outras causas comuns do >

{ sem }

' sem '

" sem "

( sem )

<<EOF sem EOF

Exemplo:

echo "teste
>


Se quiser, posso:

Analisar outros comandos curl que você estiver usando

Montar um script de backup já validado

Explicar como evitar isso usando \ ou EOF

É só falar 👍

essa resposta tá certa? curl -sS "http://localhost:9200/_snapshot/meu_backup?pretty"
{
  "meu_backup" : {
    "type" : "fs",
    "settings" : {
      "location" : "/backup"
    }
  }
}

✅ Sim, a resposta está correta.
Ela confirma que o repositório de snapshot foi registrado com sucesso no OpenSearch.

Vamos interpretar rapidamente o que cada parte significa e o que você pode validar em seguida.

📄 Resposta retornada
{
  "meu_backup" : {
    "type" : "fs",
    "settings" : {
      "location" : "/backup"
    }
  }
}

O que isso confirma

meu_backup → nome do repositório ✔

type: fs → repositório em filesystem local ✔

location: /backup → diretório onde os snapshots serão gravados ✔

Ou seja:
👉 o OpenSearch reconhece o repositório e sabe onde salvar os backups.

⚠️ Próximo check obrigatório (muito importante)

Agora você deve garantir que:

1️⃣ O diretório existe em todos os nós
ls -ld /backup

2️⃣ O usuário do OpenSearch tem permissão
chown -R opensearch:opensearch /backup
chmod 750 /backup

3️⃣ O path.repo está configurado

No opensearch.yml:

path.repo: ["/backup"]


E o OpenSearch foi reiniciado após isso.

🔎 Teste definitivo: verificar o repositório

Execute:

curl -sS "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"

✔ Resposta esperada

Algo como:

{
  "nodes" : {
    "node_id_1" : {
      "name" : "opensearch-node"
    }
  }
}


Se aparecer erro aqui, o problema é permissão ou path.repo.

🚀 Próximo passo: criar um snapshot

Exemplo simples:

curl -X PUT "http://localhost:9200/_snapshot/meu_backup/backup_teste?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "*",
    "ignore_unavailable": true
  }'

📌 Resumo
Item	Status
Repositório criado	✅ OK
Tipo fs	✅ OK
Caminho reconhecido	✅ OK
Falta validar	_verify

Se quiser, posso:

Validar a criação do snapshot

Conferir erros comuns de permissão

Montar um script de backup + rotação

É só dizer 👍

aqui está o meu arquivo de configuração do docker opensearch: # USO
# sudo docker-compose up -d
# sudo docker-compose down
# ATENÇÃO: É preciso criar uma pasta "/home/aldenor/opensearch-backup"

---
name: os

services:
  os-node1:
    image: opensearchproject/opensearch:latest
    container_name: os-node1
    environment:
      - cluster.name=os-cluster
      - node.name=os-node1
      - discovery.seed_hosts=os-node1,os-node2,os-node3
      - cluster.initial_cluster_manager_nodes=os-node1,os-node2,os-node3
      - bootstrap.memory_lock=true  
      - OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m  
      - DISABLE_SECURITY_PLUGIN=true  
      - DISABLE_INSTALL_DEMO_CONFIG=true
      - path.repo=/backup
    ulimits:
      memlock:
        soft: -1
        hard: -1
      nofile:
        soft: 65536
        hard: 65536
    volumes:
      - data1:/usr/share/opensearch/data
      - /home/aldenor/opensearch-backup:/backup
    ports:
      - "9200:9200"
      - "9601:9600"
    networks:
      - local-network-ia
    restart: always

   os-node2:
    image: opensearchproject/opensearch:latest
    container_name: os-node2
    environment:
      - cluster.name=os-cluster
      - node.name=os-node2
      - discovery.seed_hosts=os-node1,os-node2,os-node3
      - cluster.initial_cluster_manager_nodes=os-node1,os-node2,os-node3
      - bootstrap.memory_lock=true
      - OPENSEARCH_JAVA_OPTS=-Xms512m -Xmx512m      
      - DISABLE_SECURITY_PLUGIN=true  
      - DISABLE_INSTALL_DEMO_CONFIG=true
      - path.repo=/backup
    ulimits:
      memlock:
        soft: -1
        hard: -1
      nofile:
        soft: 65536
        hard: 65536
    volumes:
      - data2:/usr/share/opensearch/data
      - /home/aldenor/opensearch-backup:/backup
    networks:
      - local-network-ia
    restart: always
Pensou por 11s

Seu raciocínio está no caminho certo (montar /home/aldenor/opensearch-backup no container como /backup e definir path.repo=/backup). Mas do jeito que o compose está colado, há 3 pontos críticos que podem te derrubar:

Erro de indentação: os-node2: está com um espaço a mais (aparece como os-node2:). Isso quebra o YAML.

Você referencia os-node3 em discovery.seed_hosts e cluster.initial_cluster_manager_nodes, mas não definiu o serviço os-node3 (pelo menos no trecho enviado).

Permissões do diretório de backup no host: o OpenSearch dentro do container costuma rodar como UID 1000 (usuário opensearch). Se /home/aldenor/opensearch-backup não for gravável por esse UID, snapshot falha.

1) Como verificar se o YAML está válido (antes de subir)
docker compose -f docker-compose.yml config


Se houver problema de identação/estrutura, ele acusa.

2) Ajuste mínimo do seu compose (corrigindo identação + coerência)


Se você realmente quer 3 nós, você precisa criar o serviço os-node3 ou vai ficar com descoberta inconsistente.

3) Permissões do diretório de backup (essencial)

Crie e ajuste permissões no host:

sudo mkdir -p /home/aldenor/opensearch-backup
sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
sudo chmod -R 750 /home/aldenor/opensearch-backup


(Alternativa rápida, menos segura: chmod 777, mas prefira o chown.)

4) Subir e checar se o path.repo e o mount estão OK

Subir:

docker compose up -d


Ver se o diretório existe dentro do container:

docker exec -it os-node1 sh -lc 'ls -ld /backup && id'
docker exec -it os-node2 sh -lc 'ls -ld /backup && id'


Confirmar que o OpenSearch “enxerga” o path.repo:

curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n '"path"' -n

5) Seu repositório retornou location "/backup" — isso está correto nesse contexto?

✅ Sim, porque dentro do container o caminho é /backup.

Só falta o “teste de verdade”:

curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"


Se esse _verify falhar, quase sempre é permissão do diretório montado ou path.repo não aplicado.

Se você colar o resto do compose (principalmente se existe os-node3 e a definição da network local-network-ia), eu te devolvo a versão final já validada e consistente (2 nós ou 3 nós), pronta pra snapshot funcionar sem susto.

analise: curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n '"path"' -n
37:        "path" : {
117:        "path" : {
197:        "path" : {
aldenor@srv493387:~$ curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"
{
  "nodes" : {
    "WMdnR6ANT9WEya0koddGzw" : {
      "name" : "os-node1"
    },
    "7puV47TNQ4agKHuWkWVrUw" : {
      "name" : "os-node2"
    },
    "oCvwd9KPT965ve9Pf3gEIw" : {
      "name" : "os-node3"
    }
  }
}

Sim — pelo que você mostrou, está tudo certo e, melhor ainda, o repositório está funcional em todos os nós.

O que seus outputs provam

/_nodes/settings?pretty | grep ...
Você achou 3 ocorrências de "path" : { (linhas 37, 117, 197). Isso normalmente indica 3 nós retornando settings (cada nó aparece com seu bloco).

POST /_snapshot/meu_backup/_verify retornou:

{
  "nodes" : {
    "...": { "name": "os-node1" },
    "...": { "name": "os-node2" },
    "...": { "name": "os-node3" }
  }
}


✅ Isso é o check definitivo: o OpenSearch conseguiu acessar o repositório (/backup) a partir de os 3 containers.
Ou seja:

path.repo está aplicado

o mount /home/aldenor/opensearch-backup:/backup está acessível

permissões estão OK

o repo meu_backup está válido no cluster

Próximo passo recomendado: criar um snapshot de teste (via curl)
Snapshot de todos os índices
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/backup_teste_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "*",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo

Confirmar que foi criado
curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

Ver status (se rodar sem wait_for_completion)
curl -sS "http://localhost:9200/_snapshot/_status?pretty"

Uma melhoria no seu comando de inspeção (pra ver o path.repo claramente)

O grep que você usou só achou "path", mas não mostra o repo. Use:

curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n "path.repo" -n


ou (melhor, sem grep frágil):

curl -sS "http://localhost:9200/_nodes/settings?pretty" | sed -n '1,260p'

Observação importante

Você tem os-node3 ativo (apareceu no _verify). Então, apesar do trecho do seu compose anterior não mostrar, seu cluster de fato está com 3 nós — o que é ótimo para redundância.

Se você colar:

curl -sS "http://localhost:9200/_cat/indices?v"
e/ou

curl -sS "http://localhost:9200/_cluster/health?pretty"

eu te digo se está tudo “verde” e quais índices vale incluir no snapshot (por exemplo só os RAG/assjur) para reduzir tamanho e tempo.

curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n "path.repo" -n


Perfeito — esse output confirma 100% que o path.repo está configurado e ativo em todos os 3 nós (os-node1/2/3), e explica por que o seu grep "path.repo" não retornou nada.

1) Por que grep -n "path.repo" não achou?

Porque no JSON o OpenSearch retorna como estrutura:

"path" : {
  "repo" : [ "/backup" ]
}


Ou seja, não existe a string literal path.repo no retorno — é path → repo.

✅ O que você quer está aqui (em todos os nós):

"path" : {
  "repo" : [
    "/backup"
  ]
}

2) Como filtrar certinho só o repo (sem jq)

Você pode extrair os trechos com grep contextual:

curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n -A3 -B2 '"repo"' 


Ou pegar só as linhas com /backup:

curl -sS "http://localhost:9200/_nodes/settings?pretty" | grep -n '"/backup"'

3) Interpretação do seu output (o que está OK)

_nodes.total: 3 e successful: 3 → todos responderam.

Cada nó tem:

cluster.initial_cluster_manager_nodes = os-node1,os-node2,os-node3

discovery.seed_hosts = os-node1,os-node2,os-node3

path.repo = ["/backup"]

✅ Isso é exatamente o requisito para snapshots em cluster multi-nó: o caminho do repo tem que ser permitido em todos os nós.

4) Próximo passo: criar snapshot e verificar se gravou no host
4.1) Criar snapshot (rápido, para teste)
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/teste_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "*",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo

4.2) Confirmar no OpenSearch
curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

4.3) Confirmar no host que arquivos apareceram
ls -lah /home/aldenor/opensearch-backup
find /home/aldenor/opensearch-backup -maxdepth 2 -type f | head

5) Se quiser deixar “redondo” (boas práticas rápidas)

Para backups rotineiros, normalmente eu uso include_global_state: false (evita carregar estado global em restores por acidente).

Para restore em produção, prefira restaurar renomeando primeiro (teste), e só depois substitui.

Se você rodar o snapshot de teste e colar a saída (do PUT e do _cat/snapshots), eu confirmo se ficou SUCCESS e te digo como automatizar com retenção (apagar snapshots antigos) via shell.

curl -sS "http://localhost:9200/_cat/indices?v"
health status index                        uuid                   pri rep docs.count docs.deleted store.size pri.store.size
green  open   top_queries-2026.01.01-65730 qyhh0etzRLG_nYSPO2QCIA   1   1         63            4      283kb        141.3kb
green  open   .ql-datasources              rl5814ZSTROgEYytva1DHQ   1   2          0            0       624b           208b
green  open   top_queries-2026.01.07-65736 vJTLLlkKQ42StgXBuGFZGA   1   1         65            4      322kb        148.6kb
green  open   autos                        -89zM1LOSFacHWVipH8IkA   3   2        325           13       15mb            5mb
green  open   decisoes                     iMPu84EAQ9qUzb7g-4KNIg   3   2          0            0        2kb           684b
green  open   eventos                      UdW4ZGTWTgqjT0ODOwyQJA   3   2         36            0      4.5mb          1.5mb
green  open   autos_temp                   gXlnOhtKQMCwpmVZ6aBdqw   3   2         75            3      1.9mb        592.1kb
green  open   modelos                      5M4S0XikRDOTWaihSqSwEQ   3   2         79            0     24.3mb          8.1mb
green  open   top_queries-2026.01.04-65733 TXu41f7wQVCDf_aGbQxz9Q   1   1         57           14    573.1kb        285.9kb
green  open   .kibana_1                    ssSgKKhlSHGlnQBCWxlt4A   1   1          1            0     10.6kb          5.3kb
green  open   modelos_semantico            DbnTP1s0Q9uZbSfgS3kXqg   3   2         27            0      8.5mb          2.8mb
green  open   autos_doc_embedding          37SNXEBgQvinlRAsu9mUGA   3   2          0            0      1.8kb           624b
green  open   .opensearch-observability    1-xI58TvTKaXNkj4VNnLog   1   2          0            0       624b           208b
green  open   .plugins-ml-config           I2zfQiqRRqmRuYkdo4V2qA   1   2          1            0     12.2kb            4kb
green  open   autos_json_embedding         DNlhXBjLQvaGHEvRhxBAUA   3   2          0            0        2kb           693b
green  open   top_queries-2026.01.05-65734 oDRwgXcaSzqwqYPTEmAuVw   1   1         63            4    275.1kb        141.6kb
green  open   top_queries-2026.01.03-65732 7t5lpBy7RhSAIjusMruPnQ   1   1         63            4    266.4kb        141.5kb
green  open   rag_doc_embedding            hqo3J8JLRyCTFfWoskc2NQ   3   2          7            0      1.2mb        393.9kb
green  open   top_queries-2026.01.02-65731 VaFocK2_Q4GD2hpqJACExQ   1   1         63            4    277.9kb        138.8kb
green  open   top_queries-2026.01.06-65735 ExuF9zgPRq2pwSXB7RZAaw   1   1         63            4    283.9kb        141.7kb
aldenor@srv493387:~$ curl -sS "http://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "os-cluster",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "discovered_master" : true,
  "discovered_cluster_manager" : true,
  "active_primary_shards" : 39,
  "active_shards" : 109,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}

Ótimo — seus dois outputs mostram que o cluster está saudável e pronto para snapshot.

O que dá pra concluir (objetivamente)

/_cluster/health veio green, com 3 nós e 0 shards não atribuídos. Isso é o cenário ideal para backup.

/_cat/indices também está todo green/open. Seus índices “relevantes do sistema” (no seu caso) parecem ser: autos, autos_temp, eventos, modelos, modelos_semantico, rag_doc_embedding e possivelmente os embeddings autos_doc_embedding / autos_json_embedding (ainda estão vazios).

Recomendo 2 tipos de backup (curl pronto)
1) Snapshot “dados do seu app” (sem estado global)

Esse é o que você roda todo dia/semana, sem carregar coisas do cluster (.kibana, observability etc.):

curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/app_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo


Depois confira:

curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

2) Snapshot “full” (inclui estado global)

Esse você roda com menos frequência (por exemplo semanal/mensal), porque inclui configurações globais do cluster:

curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/full_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "*",
    "ignore_unavailable": true,
    "include_global_state": true
  }'
echo

Confirmar que o snapshot foi gravado no host

Como você montou /home/aldenor/opensearch-backup:/backup, os arquivos devem aparecer aí:

ls -lah /home/aldenor/opensearch-backup
find /home/aldenor/opensearch-backup -maxdepth 2 -type f | head -n 50

Observação útil sobre índices “.kibana / .opensearch / .plugins”

Se você não precisa restaurar dashboards/configs do OpenSearch Dashboards, mantenha include_global_state:false e não inclua índices começando com . no backup “app”.

Se você usa Dashboards e quer preservar, faça o snapshot “full” ocasionalmente.

Se quiser, te deixo um comando de “limpeza” (apagar snapshots antigos)

Exemplo: apagar um snapshot específico:

curl -sS -X DELETE "http://localhost:9200/_snapshot/meu_backup/NOME_DO_SNAPSHOT?pretty"
echo


Se você me disser sua política (ex.: “manter 7 diários e 4 semanais”), eu te monto um script shell que lista snapshots por data e apaga o excedente.



O que aconteceu

O snapshot falhou por permissão no diretório do repositório:

access_denied_exception tentando escrever em:

/backup/indices/.../meta-....dat

e também “failed to create blob container” em /backup/indices/...

Isso significa: o processo do OpenSearch dentro do container não tem permissão de escrita no diretório montado /backup (que no host é /home/aldenor/opensearch-backup).

✅ O _verify ter dado OK só prova que ele consegue acessar o path, não que ele consegue criar/alterar todos os arquivos necessários (na prática, o snapshot é o teste real de escrita).

E o fato de existir um snapshot antigo snapshot_1 SUCCESS indica que ou as permissões mudaram depois, ou esse snapshot foi feito em outro cenário (outro diretório, outro usuário, ou antes de alguma mudança de ownership).

1) Confirme o usuário (UID/GID) que o OpenSearch usa dentro do container

Rode em cada nó:

docker exec -it os-node1 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'
docker exec -it os-node2 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'
docker exec -it os-node3 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'


Se der Permission denied no touch, bate exatamente com o erro do snapshot.

2) Corrija permissões no HOST (a solução mais comum)

No host:

sudo mkdir -p /home/aldenor/opensearch-backup
sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
sudo chmod -R 750 /home/aldenor/opensearch-backup


Na imagem oficial, o opensearch costuma rodar como UID 1000. Se no seu id dentro do container aparecer outro UID, use ele no chown.

Depois, reinicie os containers (pra garantir que não ficou nada travado):

docker compose restart

3) Se ainda falhar: pode ser “root-squash” / FS com restrição / SELinux/AppArmor

Isso pode acontecer se /home estiver em FS com política mais restritiva, ou SELinux (menos comum no Ubuntu, mais em CentOS/RHEL).

3.1) Verifique permissões e dono no host
ls -ld /home/aldenor/opensearch-backup
ls -la /home/aldenor/opensearch-backup | head

3.2) Teste com permissão ampla (só pra diagnosticar)

⚠️ Só para teste rápido (depois volte):

sudo chmod -R 777 /home/aldenor/opensearch-backup


Se com 777 funcionar, é 100% permissão/ownership.

4) Tente o snapshot de novo (após corrigir)
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/app_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo


E confira:

curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

5) Por que o erro menciona /backup/indices/<UUID>/meta-...dat?

O layout do repositório fs cria uma árvore tipo:

/backup/index-*

/backup/indices/<index_uuid>/...

arquivos meta-*.dat e snap-*.dat

Se o OpenSearch não consegue criar/escrever ali, o snapshot não fecha.

6) Extra: deixe isso “blindado” no docker-compose

Se quiser evitar dor com UID no futuro, você pode fixar o usuário do container para bater com o dono do diretório (depende do seu ambiente). Exemplo (apenas se necessário):

user: "1000:1000"


Mas eu só recomendo depois de confirmar com docker exec ... id qual UID ele realmente usa.

O que eu faria agora, na sua ordem

Rodar o docker exec ... touch /backup/.write_test (vai confirmar em 10s).

Ajustar chown do diretório no host pro UID correto (provavelmente 1000).

Repetir o snapshot.

Se você colar o resultado de docker exec -it os-node1 sh -lc 'id; ls -ld /backup', eu te digo exatamente qual chown usar (sem chute).

docker exec -it os-node1 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'
docker exec -it os-node2 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'
docker exec -it os-node3 sh -lc 'id && ls -ld /backup && touch /backup/.write_test && ls -l /backup/.write_test && rm -f /backup/.write_test'
uid=1000(opensearch) gid=1000(opensearch) groups=1000(opensearch)
drwxrwxrwx 4 1002 1002 4096 Jan  7 21:48 /backup
-rw-r--r-- 1 opensearch opensearch 0 Jan  7 21:50 /backup/.write_test
uid=1000(opensearch) gid=1000(opensearch) groups=1000(opensearch)
drwxrwxrwx 4 1002 1002 4096 Jan  7 21:50 /backup
-rw-r--r-- 1 opensearch opensearch 0 Jan  7 21:50 /backup/.write_test
uid=1000(opensearch) gid=1000(opensearch) groups=1000(opensearch)
drwxrwxrwx 4 1002 1002 4096 Jan  7 21:50 /backup
-rw-r--r-- 1 opensearch opensearch 0 Jan  7 21:50 /backup/.write_test

Excelente teste — e ele revela exatamente a causa do seu erro de snapshot:

O OpenSearch roda como uid=1000(opensearch) gid=1000(opensearch).

O diretório /backup está com dono 1002:1002 (veja: drwxrwxrwx 4 1002 1002 ... /backup).

Você conseguiu criar /backup/.write_test, mas snapshot falhou com access_denied ao criar arquivos em /backup/indices/....

Isso acontece muito quando:

O diretório raiz tem permissão, mas subpastas/arquivos já existentes (como /backup/indices/...) ficaram com dono/permissão incompatíveis (ou com ACL/atributos especiais).

Ou houve criação prévia dessas pastas por outro usuário (UID 1002) e agora o processo 1000 não consegue atualizar certos arquivos (meta-*.dat) mesmo com 777 no topo (pode haver arquivos com permissão restrita, ou atributos imutáveis, ou ACL).

Mesmo com drwxrwxrwx, se algum arquivo dentro estiver, por exemplo, -rw------- e dono 1002, o opensearch (1000) não consegue sobrescrever.

✅ Correção definitiva (no HOST): normalizar ownership/perms recursivamente

Como o OpenSearch roda com UID/GID 1000, deixe o repo inteiro pertencendo a 1000:1000.

No host (fora do container):

sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
sudo find /home/aldenor/opensearch-backup -type d -exec chmod 775 {} \;
sudo find /home/aldenor/opensearch-backup -type f -exec chmod 664 {} \;


Se você prefere manter simples (e você já está com 777), faça pelo menos o chown -R 1000:1000, que é o que resolve quase sempre.

Depois, reinicie os nós (não é sempre obrigatório, mas ajuda a evitar handle antigo):

docker compose restart

🔍 Verificação rápida (garantir que não ficou nada com UID 1002)
sudo find /home/aldenor/opensearch-backup -not -user 1000 -o -not -group 1000 | head -n 50


Se aparecer coisa, é exatamente isso que quebrava o snapshot.

🧹 Se você quiser “zerar” o repositório (opcional, mas costuma ser o jeito mais limpo)

Se você não precisa manter o snapshot_1 e quer eliminar qualquer sujeira/permissão estranha:

Apague o conteúdo do repo no host:

sudo rm -rf /home/aldenor/opensearch-backup/*
sudo chown -R 1000:1000 /home/aldenor/opensearch-backup


(Opcional) Recrie o repo no OpenSearch (não é obrigatório se o nome é o mesmo, mas fica limpo):

curl -sS -X DELETE "http://localhost:9200/_snapshot/meu_backup?pretty"
echo
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup" \
  -H "Content-Type: application/json" \
  -d '{"type":"fs","settings":{"location":"/backup","compress":true}}'
echo


Verifique:

curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"

✅ Agora rode o snapshot de novo
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/app_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo


E confira:

curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"

🎯 Por que seu touch funcionou, mas snapshot não?

Porque touch /backup/.write_test cria um arquivo novo no diretório raiz, onde você tem permissão.
O snapshot estava tentando escrever em caminhos já existentes (ex.: /backup/indices/.../meta-*.dat) que podem estar com dono/permissão/ACL que bloqueia a atualização.

Se você rodar o find ... -not -user 1000 e colar as primeiras linhas (ou o resultado do snapshot após o chown -R), eu confirmo que ficou 100% resolvido.

sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
sudo find /home/aldenor/opensearch-backup -type d -exec chmod 775 {} \;
sudo find /home/aldenor/opensearch-backup -type f -exec chmod 664 {} \;
[sudo] password for aldenor: 
aldenor@srv493387:~$ sudo find /home/aldenor/opensearch-backup -not -user 1000 -o -not -group 1000 | head -n 50
aldenor@srv493387:~$ sudo rm -rf /home/aldenor/opensearch-backup/*
sudo chown -R 1000:1000 /home/aldenor/opensearch-backup
aldenor@srv493387:~$ curl -sS -X DELETE "http://localhost:9200/_snapshot/meu_backup?pretty"
echo
curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup" \
  -H "Content-Type: application/json" \
  -d '{"type":"fs","settings":{"location":"/backup","compress":true}}'
echo
{
  "acknowledged" : true
}

{"acknowledged":true}
aldenor@srv493387:~$ curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/_verify?pretty"
{
  "nodes" : {
    "WMdnR6ANT9WEya0koddGzw" : {
      "name" : "os-node1"
    },
    "7puV47TNQ4agKHuWkWVrUw" : {
      "name" : "os-node2"
    },
    "oCvwd9KPT965ve9Pf3gEIw" : {
      "name" : "os-node3"
    }
  }
}
aldenor@srv493387:~$ curl -sS -X PUT "http://localhost:9200/_snapshot/meu_backup/app_$(date +%F_%H%M)?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "autos,autos_temp,eventos,modelos,modelos_semantico,rag_doc_embedding,decisoes,autos_doc_embedding,autos_json_embedding",
    "ignore_unavailable": true,
    "include_global_state": false
  }'
echo
{"snapshot":{"snapshot":"app_2026-01-07_1853","uuid":"vKlt1a2nTSS25GAsB8p59Q","version_id":136407927,"version":"2.19.1","remote_store_index_shallow_copy":false,"indices":["rag_doc_embedding","modelos","autos_doc_embedding","autos","autos_temp","eventos","modelos_semantico","decisoes","autos_json_embedding"],"data_streams":[],"include_global_state":false,"state":"SUCCESS","start_time":"2026-01-07T21:53:21.870Z","start_time_in_millis":1767822801870,"end_time":"2026-01-07T21:53:23.472Z","end_time_in_millis":1767822803472,"duration_in_millis":1602,"failures":[],"shards":{"total":27,"failed":0,"successful":27}}}
aldenor@srv493387:~$ curl -sS "http://localhost:9200/_cat/snapshots/meu_backup?v"
id                   status start_epoch start_time end_epoch  end_time duration indices successful_shards failed_shards total_shards
app_2026-01-07_1853 SUCCESS 1767822801  21:53:21   1767822803 21:53:23     1.6s       9                27             0           27

Perfeito — agora está 100% resolvido e seu procedimento ficou correto e reproduzível.

O que você fez foi exatamente o “gold standard” pra consertar repo fs com permissão/ownership bagunçados:

Normalizou permissões/ownership ✅

Limpou o diretório do repo ✅

Recriou o repositório no OpenSearch ✅

Verificou (_verify) ✅

Criou snapshot com state: "SUCCESS" ✅

Confirmou no _cat/snapshots ✅

✅ Evidências que está tudo OK

Resposta do snapshot:

"state":"SUCCESS"

"failures":[]

"shards":{"total":27,"failed":0,"successful":27}

_cat/snapshots:

app_2026-01-07_1853 SUCCESS

Isso significa: backup criado com sucesso no repositório meu_backup e gravado no seu diretório montado.

Próximos comandos úteis (pra você já deixar pronto)
1) Conferir que os arquivos foram gerados no host
ls -lah /home/aldenor/opensearch-backup
find /home/aldenor/opensearch-backup -maxdepth 2 -type f | head -n 50

2) Ver detalhes do snapshot (lista de índices, etc.)
curl -sS "http://localhost:9200/_snapshot/meu_backup/app_2026-01-07_1853?pretty"

3) Restore de teste RENOMEANDO (recomendado)

Ex.: restaurar rag_doc_embedding como rag_doc_embedding_restore_test

curl -sS -X POST "http://localhost:9200/_snapshot/meu_backup/app_2026-01-07_1853/_restore?wait_for_completion=true" \
  -H "Content-Type: application/json" \
  -d '{
    "indices": "rag_doc_embedding",
    "rename_pattern": "rag_doc_embedding",
    "rename_replacement": "rag_doc_embedding_restore_test",
    "include_global_state": false
  }'
echo


Depois:

curl -sS "http://localhost:9200/_cat/indices?v" | grep rag_doc_embedding

Automação simples (backup “app” com nome por data)

Você já tem o comando. Se quiser deixar como script:

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


Rodar:

./backup_opensearch_app.sh

Nota importante (pra não voltar o problema)

Como o OpenSearch roda como UID 1000, mantenha o diretório do repo no host sempre com dono 1000:1000:

sudo chown -R 1000:1000 /home/aldenor/opensearch-backup


Se você me disser sua política de retenção (ex.: “manter 7 backups app e 4 full”), eu te passo um script bash que apaga snapshots antigos automaticamente com curl (sem depender de jq).

O ChatGPT pode cometer erros. Confira informações importantes. Consulte as Preferências de cookies.
