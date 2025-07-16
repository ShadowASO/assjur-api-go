package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonPeticao(doc PeticaoDiversa) string {
	var sb strings.Builder

	//Natureza do documento: Contestação
	sb.WriteString(doc.Tipo.Description + ": ")

	// Causa de pedir
	sb.WriteString("Causa de Pedir: " + doc.CausaDePedir)

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("\nPedidos: ")
		for _, p := range doc.Pedidos {
			sb.WriteString(p)
		}
	}

	return sb.String()
}

func ParserPeticaoJson(idNatu int, docJson json.RawMessage) (string, error) {

	var doc PeticaoDiversa
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonPeticao(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
