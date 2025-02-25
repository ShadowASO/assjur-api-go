#  Index 'modelos'
{
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