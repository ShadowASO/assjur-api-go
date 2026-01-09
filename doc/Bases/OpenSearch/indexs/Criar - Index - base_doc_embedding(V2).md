PUT /base_doc_embedding
{
  "settings": {
    "index.knn": true,
    "number_of_shards": 3,
    "number_of_replicas": 2,
    "analysis": {
      "analyzer": {
        "brazilian": { "type": "brazilian" }
      }
    }
  },
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "id_ctxt": { "type": "keyword" },
      "id_pje":  { "type": "keyword" },
      "hash_texto": { "type": "keyword" },
      "username_inc": { "type": "keyword", "ignore_above": 256 },
      "dt_inc": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "status": { "type": "keyword", "ignore_above": 16 },
      "classe": {
        "type": "text",
        "analyzer": "brazilian",
        "fields": {
          "kw": { "type": "keyword", "ignore_above": 256 }
        }
      },
      "assunto": {
        "type": "text",
        "analyzer": "brazilian",
        "fields": {
          "kw": { "type": "keyword", "ignore_above": 256 }
        }
      },
      "natureza": {
        "type": "text",
        "analyzer": "brazilian",
        "fields": {
          "kw": { "type": "keyword", "ignore_above": 256 }
        }
      },
      "tipo": {
        "type": "text",
        "analyzer": "brazilian",
        "fields": {
          "kw": { "type": "keyword", "ignore_above": 256 }
        }
      },

      "tema": {
        "type": "text",
        "analyzer": "brazilian",
        "fields": {
          "kw": { "type": "keyword", "ignore_above": 256 }
        }
      },

      "fonte": {
        "type": "keyword",
        "ignore_above": 512,
        "fields": {
          "text": { "type": "text", "analyzer": "brazilian" }
        }
      },

      "texto": {
        "type": "text",
        "analyzer": "brazilian"
      },

      
      "texto_embedding": {
        "type": "knn_vector",
        "dimension": 3072,
        "method": {
          "name": "hnsw",
          "space_type": "cosinesimil",
          "engine": "faiss",
          "parameters": {
            "m": 16,
            "ef_construction": 128
          }
        }
      }
    }
  }
}

