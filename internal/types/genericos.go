package types

// Tipo para construir objetos JSON
// Exemplo:
//
//	query := JsonMap{
//	    "size": 10,
//	    "query": JsonMap{
//	        "terms": JsonMap{
//	            "id_ctxt": []int{123, 456, 789},
//	        },
//	    },
//	}
type JsonMap map[string]interface{}
