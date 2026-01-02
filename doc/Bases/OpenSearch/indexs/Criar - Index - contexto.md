PUT /contexto
{
  "settings": {
    "number_of_shards": 3,
    "number_of_replicas": 2
  },
  "mappings": {
    "dynamic": "strict",
    "properties": {
      "id_ctxt": { "type": "keyword" },

      "nr_proc": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 32 }
        }
      },

      "juizo": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 256 }
        }
      },

      "classe": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 256 }
        }
      },

      "assunto": {
        "type": "text",
        "fields": {
          "keyword": { "type": "keyword", "ignore_above": 256 }
        }
      },

      "prompt_tokens": { "type": "integer" },
      "completion_tokens": { "type": "integer" },

      "dt_inc": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "username_inc": { "type": "keyword", "ignore_above": 20  },

      "status": { "type": "keyword", "ignore_above": 1 }
    }
  }
}

