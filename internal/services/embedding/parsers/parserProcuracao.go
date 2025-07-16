package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// formatarJsonProcuracao monta texto para embedding a partir do modelo padronizado de Procuração
func formatarJsonProcuracao(doc Procuracao) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	// Outorgantes
	if len(doc.Outorgantes) > 0 {
		sb.WriteString("Outorgantes: ")
		for _, p := range doc.Outorgantes {
			sb.WriteString(p.Nome + "; ")
		}
	}

	// Poderes
	sb.WriteString("Poderes: " + doc.Poderes + "; ")

	// Advogados
	if len(doc.Advogados) > 0 {
		sb.WriteString("Advogados: ")
		for _, adv := range doc.Advogados {
			sb.WriteString(adv.Nome + " (" + adv.OAB + "); ")
		}
	}

	return sb.String()
}

// ParserProcuracaoJson desserializa e formata JSON do tipo Procuração
func ParserProcuracaoJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc Procuracao
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON da procuração: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da procuração")
	}
	textoFormatado := formatarJsonProcuracao(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
