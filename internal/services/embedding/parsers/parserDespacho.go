package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

type Despacho struct {
	Tipo       Tipo         `json:"tipo"`
	Processo   string       `json:"processo"`
	IdPje      string       `json:"id_pje"`
	Conteudo   Conteudo     `json:"conteudo"`
	Deliberado []Deliberado `json:"deliberado"`
	Juiz       Juiz         `json:"juiz"`
}

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonDespachos(doc Despacho) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	// Conteudo
	if len(doc.Conteudo) > 0 {
		//sb.WriteString("Conteúdo:\n" + "\n")
		for _, v := range doc.Conteudo {
			sb.WriteString(v + ";")
		}
	}

	// Deliberado
	if len(doc.Deliberado) > 0 {
		sb.WriteString("\nDeliberações:")
		for _, v := range doc.Deliberado {
			sb.WriteString(v.Finalidade + ";")
			sb.WriteString(v.Destinatario + ";")
			sb.WriteString(v.Prazo + "\n")
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
