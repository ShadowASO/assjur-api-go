# PUT /mni_documentos_tmp
{
  "settings": {
    "index": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "refresh_interval": "10s"
    }
  },
  "mappings": {
    "dynamic": false,
    "properties": {
      "id_ctxt": {
        "type": "keyword"
      },
      "numero_processo": {
        "type": "keyword"
      },
      "id_pje": {
        "type": "keyword"
      },
      "tipo_documento": {
        "type": "keyword"
      },
      "descricao": {
        "type": "keyword"
      },
      "storage_id": {
        "type": "keyword"
      },
      "mimetype": {
        "type": "keyword"
      },
      "formato_entrega": {
        "type": "keyword"
      },
      "conteudo_texto": {
        "type": "text",
        "index": false
      },
      "conteudo_html": {
        "type": "text",
        "index": false
      },
      "conteudo_base64": {
        "type": "binary"
      },
      "status": {
        "type": "keyword"
      },
      "erro": {
        "type": "text",
        "index": false
      }, 
      "data_juntada_millis": {
        "type": "long"
      },
      "usuario_juntada_arquivo": {
        "type": "keyword"
      },
      "criado_em": {
        "type": "date"
      },
      "atualizado_em": {
        "type": "date"
      },
      "expira_em": {
        "type": "date"
      }
    }
  }
}
