PUT /rag_doc_embedding
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
      "id_pje": {
        "type": "keyword"
      },
      "classe": {
        "type": "keyword"
      },
      "assunto": {
        "type": "keyword"
      },
      "natureza": {
        "type": "keyword"
      },
      "tipo": {
        "type": "keyword"
      },
      "tema": {
        "type": "text",
        "analyzer": "brazilian"
      },
      "fonte": {
        "type": "text",
        "analyzer": "brazilian"
      },
      "data_texto": {
        "type": "text",
        "analyzer": "brazilian"
      },
      "data_embedding": {
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
