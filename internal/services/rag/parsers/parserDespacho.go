package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/mserror"
	"ocrserver/internal/utils/mslogger"

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
	mslogger.LoggerGlobal.Info("Entrei")
	var doc Despacho
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		mslogger.LoggerGlobal.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", mserror.NewError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonDespachos(doc)
	//mslogger.LoggerGlobal.Info(textoFormatado)
	return textoFormatado, nil
}
