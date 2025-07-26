package consts

import ()

type AutosRow struct {
	IdCtxt       int       `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	IdPje        string    `json:"id_pje"`
	Doc          string    `json:"doc"`
	DocJsonRaw   string    `json:"doc_json_raw"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

type ResponseAutosRow struct {
	Id           string    `json:"id"`
	IdCtxt       int       `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	IdPje        string    `json:"id_pje"`
	Doc          string    `json:"doc"`
	DocJsonRaw   string    `json:"doc_json_raw"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

type AutosTempRow struct {
	//Id     string `json:"id"`
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

type ResponseAutosTempRow struct {
	Id     string `json:"id"`
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

type AutosJsonEmbeddingRow struct {
	IdDoc        string    `json:"id_doc"`
	IdCtxt       int       `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	DocEmbedding []float32 `json:"doc_embedding"`
}
type ResponseAutosJsonEmbeddingRow struct {
	Id           string    `json:"id"`
	IdDoc        string    `json:"id_doc"`
	IdCtxt       int       `json:"id_ctxt"`
	IdNatu       int       `json:"id_natu"`
	DocEmbedding []float32 `json:"doc_embedding"`
}

// Converte campo object dos indices do OpenSearch para json.RawMessage
// func ParseObjectOpensearchToRawMessage(docJsonRaw string) (json.RawMessage, error) {
// 	var rawJson json.RawMessage
// 	var err error
// 	if docJsonRaw != nil {
// 		rawJson, err = json.Marshal(docJsonRaw)
// 		if err != nil {
// 			logger.Log.Errorf("Erro ao serializar docJson para json.RawMessage: %v", err)
// 			return nil, erros.CreateError("Erro ao serializar docJson para json.RawMessage: %v", err.Error())
// 		}
// 	}
// 	return rawJson, nil
// }
