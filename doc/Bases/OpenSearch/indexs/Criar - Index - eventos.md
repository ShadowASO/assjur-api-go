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
      "username_inc": { "type": "keyword", "ignore_above": 256 },
      "dt_inc": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
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
