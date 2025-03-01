#  Index 'modelos'
{
    "settings": {
        "number_of_shards":3,
        "number_of_replicas": 2
    },
	"mappings": {
		"properties": {
			"natureza": {
				"type": "text",
				"fields": {
					"keyword": {
						"type": "keyword"
					}
				}
			},
			"ementa": {
				"type": "text"
			},
			"inteiro_teor": {
				"type": "text"
			}
		}
	}
}
