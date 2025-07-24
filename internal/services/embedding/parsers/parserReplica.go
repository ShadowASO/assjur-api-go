package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// formatarJsonReplica monta texto para embedding a partir do modelo padronizado
func formatarJsonReplica(doc Replica) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	sb.WriteString("Fatos: " + doc.Fatos + "; ")

	if len(doc.QuestoesControvertidas) > 0 {
		sb.WriteString("Questões Controvertidas: ")
		for _, v := range doc.QuestoesControvertidas {
			sb.WriteString(v + "; ")
		}
	}

	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos: ")
		for _, p := range doc.Pedidos {
			sb.WriteString(p + "; ")
		}
	}

	return sb.String()
}

// ParserReplicaJson deserializa e formata JSON do tipo Réplica
func ParserReplicaJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc Replica
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON da réplica: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da réplica")
	}
	textoFormatado := formatarJsonReplica(doc)
	//logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
