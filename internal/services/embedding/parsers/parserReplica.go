package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

type Replica struct {
	Tipo                   Tipo       `json:"tipo"`
	Processo               string     `json:"processo"`
	IdPje                  string     `json:"id_pje"`
	Peticionante           []Parte    `json:"peticionante"`
	Fatos                  string     `json:"fatos"`
	QuestoesControvertidas []string   `json:"questoes_controvertidas"`
	Pedidos                []string   `json:"pedidos"`
	Provas                 []string   `json:"provas"`
	RolTestemunhas         []string   `json:"rol_testemunhas"`
	Advogados              []Advogado `json:"advogados"`
}

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonReplica(doc Replica) string {
	var sb strings.Builder

	//Natureza do documento: Contestação
	sb.WriteString(doc.Tipo.Description + ": ")

	// Fatos
	sb.WriteString("Fatos: " + doc.Fatos + "; ")

	// Questões controvertidas
	if len(doc.QuestoesControvertidas) > 0 {
		sb.WriteString("Questões Controvertidas: ")
		for _, v := range doc.QuestoesControvertidas {
			sb.WriteString(v + "; ")
		}

	}

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos: ")
		for _, p := range doc.Pedidos {
			sb.WriteString(p + "; ")
		}
	}

	return sb.String()
}

func ParserReplicaJson(idNatu int, docJson json.RawMessage) (string, error) {

	var doc Replica
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonReplica(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
