### Criando os novos campos:
PUT /modelos/_mapping
{
  "properties": {
    "ementa_html": {
      "type": "text",
      "index": false
    },
    "inteiro_teor_html": {
      "type": "text",
      "index": false
    }
  }
}

### Formato final do índex:
PUT /modelos
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
        "ementa": {
          "type": "text",
          "analyzer": "brazilian"
        },
        "ementa_embedding": {
          "type": "knn_vector",
          "dimension": 3072,
          "method": {
            "engine": "faiss",
            "space_type": "cosinesimil",
            "name": "hnsw",
            "parameters": {}
          }
        },
        "ementa_html": {
          "type": "text",
          "index": false
        },
        "inteiro_teor": {
          "type": "text",
          "analyzer": "brazilian"
        },
        "inteiro_teor_embedding": {
          "type": "knn_vector",
          "dimension": 3072,
          "method": {
            "engine": "faiss",
            "space_type": "cosinesimil",
            "name": "hnsw",
            "parameters": {}
          }
        },
        "inteiro_teor_html": {
          "type": "text",
          "index": false
        },
        "natureza": {
          "type": "keyword"
        }
      }
    }
  }


