package consts

import "time"

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
	IdCtxt int       `json:"id_ctxt"`
	IdNatu int       `json:"id_natu"`
	IdPje  string    `json:"id_pje"`
	DtInc  time.Time `json:"dt_inc"` // data/hora da inclusão
	Doc    string    `json:"doc"`
}

type ResponseAutosTempRow struct {
	Id     string    `json:"id"`
	IdCtxt int       `json:"id_ctxt"`
	IdNatu int       `json:"id_natu"`
	IdPje  string    `json:"id_pje"`
	DtInc  time.Time `json:"dt_inc"` // data/hora da inclusão
	Doc    string    `json:"doc"`
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
