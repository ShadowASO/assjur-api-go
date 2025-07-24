package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// formatarJsonContestacao monta texto para embedding a partir do modelo padronizado
func formatarJsonContestacao(doc Contestacao) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	sb.WriteString("Fatos: " + doc.Fatos + "; ")

	if len(doc.Preliminares) > 0 {
		sb.WriteString("Preliminares: ")
		for _, v := range doc.Preliminares {
			sb.WriteString(v + "; ")
		}
	}

	if len(doc.AtosNormativos) > 0 {
		sb.WriteString("Atos Normativos: ")
		for _, v := range doc.AtosNormativos {
			sb.WriteString(v + "; ")
		}
	}

	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos: ")
		for _, v := range doc.Pedidos {
			sb.WriteString(v + "; ")
		}
	}

	sb.WriteString("Tutela Provisória: " + doc.TutelaProvisoria.Detalhes + "; ")

	if len(doc.QuestoesControvertidas) > 0 {
		sb.WriteString("Questões Controvertidas: ")
		for _, v := range doc.QuestoesControvertidas {
			sb.WriteString(v + "; ")
		}
	}

	return sb.String()
}

// ParserContestacaoJson deserializa e formata JSON da contestação
func ParserContestacaoJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc Contestacao
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON da contestação: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da contestação")
	}
	textoFormatado := formatarJsonContestacao(doc)
	//logger.Log.Infof(textoFormatado)
	return textoFormatado, nil
}
