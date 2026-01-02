PUT /eventos
{
  "settings": {
    "index.knn": true,
    "number_of_shards": 3,
    "number_of_replicas": 2,
    "analysis": {
      "analyzer": {
        "brazilian": {
          "type": "brazilian"
        }
      }
    }
  },
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "id_ctxt": { "type": "keyword" },
      "id_natu": {
        "type": "integer"
      },
      "id_pje": {
        "type": "keyword",
        "ignore_above": 20
      },
      "doc": {
        "type": "text",
        "analyzer": "brazilian"
      },
      "doc_json_raw": {
        "type": "keyword",
        "ignore_above": 100000
      },
      "doc_embedding": {
        "type": "knn_vector",
        "dimension": 3072,
        "method": {
          "name": "hnsw",
          "space_type": "cosinesimil",
          "engine": "faiss"
        }
      }
    }
  }
}
