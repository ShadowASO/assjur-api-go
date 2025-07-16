package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonDespachos(doc Despacho) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	// Conteudo
	if len(doc.Conteudo) > 0 {
		for _, v := range doc.Conteudo {
			sb.WriteString(v)
		}
	}

	// Deliberado
	if len(doc.Deliberado) > 0 {
		sb.WriteString("\nDeliberações:")
		for _, v := range doc.Deliberado {
			sb.WriteString("\nfinalidade: " + v.Finalidade + ". ")
			sb.WriteString("destinatário: " + v.Destinatario + ". ")
			sb.WriteString("prazo: " + v.Prazo + ";")
		}
	}

	return sb.String()
}

func ParserDespachoJson(idNatu int, docJson json.RawMessage) (string, error) {
	logger.Log.Info("Entrei")
	var doc Despacho
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonDespachos(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
