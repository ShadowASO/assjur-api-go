package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// formatarJsonRolTestemunhas monta texto para embedding a partir do modelo padronizado de Rol de Testemunhas
func formatarJsonRolTestemunhas(doc RolTestemunhas) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	// Partes que arrolaram testemunhas
	if len(doc.Partes) > 0 {
		sb.WriteString("Partes: ")
		for _, p := range doc.Partes {
			sb.WriteString(p.Nome + "; ")
		}
	}

	// Testemunhas
	if len(doc.Testemunhas) > 0 {
		sb.WriteString("Testemunhas: ")
		for _, t := range doc.Testemunhas {
			sb.WriteString(t.Nome + "; ")
		}
	}

	// Advogados
	if len(doc.Advogados) > 0 {
		sb.WriteString("Advogados: ")
		for _, adv := range doc.Advogados {
			sb.WriteString(adv.Nome + " (" + adv.OAB + "); ")
		}
	}

	return sb.String()
}

// ParserRolTestemunhasJson desserializa e formata JSON do tipo Rol de Testemunhas
func ParserRolTestemunhasJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc RolTestemunhas
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON do rol de testemunhas: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON do rol de testemunhas")
	}
	textoFormatado := formatarJsonRolTestemunhas(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
