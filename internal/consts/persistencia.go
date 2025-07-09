package consts

import (
	"encoding/json"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

type AutosRow struct {
	Id           string                 `json:"id"`
	IdCtxt       int                    `json:"id_ctxt"`
	IdNatu       int                    `json:"id_natu"`
	IdPje        string                 `json:"id_pje"`
	Doc          string                 `json:"doc"`
	DocJson      map[string]interface{} `json:"doc_json"`
	DocEmbedding []float32              `json:"doc_embedding"`
}

type Autos_tempRow struct {
	Id     string `json:"id"`
	IdCtxt int    `json:"id_ctxt"`
	IdNatu int    `json:"id_natu"`
	IdPje  string `json:"id_pje"`
	Doc    string `json:"doc"`
}

// Converte campo object dos indices do OpenSearch para json.RawMessage
func ParseObjectOpensearchToRawMessage(docJson map[string]interface{}) (json.RawMessage, error) {
	var rawJson json.RawMessage
	var err error
	if docJson != nil {
		rawJson, err = json.Marshal(docJson)
		if err != nil {
			logger.Log.Errorf("Erro ao serializar docJson para json.RawMessage: %v", err)
			return nil, erros.CreateError("Erro ao serializar docJson para json.RawMessage: %v", err.Error())
		}
	}
	return rawJson, nil
}
