PUT /autos_temp
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
      "id_ctxt": {
        "type": "integer"
      },
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
      }
    }
  }
}
