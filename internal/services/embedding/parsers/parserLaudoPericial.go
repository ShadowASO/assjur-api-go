package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// formatarJsonLaudoPericial monta texto para embedding a partir do modelo padronizado de Laudo Pericial
func formatarJsonLaudoPericial(doc LaudoPericial) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	// Peritos
	if len(doc.Peritos) > 0 {
		sb.WriteString("Peritos: ")
		for _, p := range doc.Peritos {
			sb.WriteString(p.Nome + "; ")
		}
	}

	// Conclusões
	sb.WriteString("Conclusões: " + doc.Conclusoes + "; ")

	return sb.String()
}

// ParserLaudoPericialJson desserializa e formata JSON do tipo Laudo Pericial
func ParserLaudoPericialJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc LaudoPericial
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON do laudo pericial: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON do laudo pericial")
	}
	textoFormatado := formatarJsonLaudoPericial(doc)
	//logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
