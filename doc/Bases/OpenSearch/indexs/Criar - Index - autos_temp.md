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
      "id_ctxt": { "type": "keyword" },
      "id_natu": {
        "type": "integer"
      },
      "id_pje": {
        "type": "keyword",
        "ignore_above": 20
      },
      "dt_inc": {
      "type": "date",
      "format": "strict_date_optional_time||epoch_millis"
      },
      "doc": {
        "type": "text",
        "analyzer": "brazilian"
      }
    }
  }
}
