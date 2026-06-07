package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/mserror"
	"ocrserver/internal/utils/mslogger"
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
		mslogger.LoggerGlobal.Errorf("Erro ao realizar Unmarshal do JSON da contestação: ", err)
		return "", mserror.NewError("Erro ao realizar Unmarshal de JSON da contestação")
	}
	textoFormatado := formatarJsonContestacao(doc)
	//mslogger.LoggerGlobal.Infof(textoFormatado)
	return textoFormatado, nil
}
