package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonSentenca(doc Sentenca) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	sb.WriteString("\nProcesso: " + doc.Processo)
	sb.WriteString("\nID PJE: " + doc.IdPje)

	// Preliminares
	if len(doc.Preliminares) > 0 {
		sb.WriteString("\nPreliminares:\n")
		for _, p := range doc.Preliminares {
			sb.WriteString("- " + p.Assunto + ": " + p.Decisao + "\n")
		}
	}

	// Fundamentos
	if len(doc.Fundamentos) > 0 {
		sb.WriteString("\nFundamentos:\n")
		for _, f := range doc.Fundamentos {
			sb.WriteString(f.Texto + "\n")
			if len(f.Provas) > 0 {
				sb.WriteString("Provas: ")
				sb.WriteString(strings.Join(f.Provas, ", ") + "\n")
			}
		}
	}

	// Conclusão
	if len(doc.Conclusao) > 0 {
		sb.WriteString("\nConclusão:\n")
		for _, c := range doc.Conclusao {
			sb.WriteString("- Resultado: " + c.Resultado + "\n")
			sb.WriteString("  Destinatário: " + c.Destinatario + "\n")
			sb.WriteString("  Prazo: " + c.Prazo + "\n")
			sb.WriteString("  Decisão: " + c.Decisao + "\n")
		}
	}

	return sb.String()
}

func ParserSentencaJson(idNatu int, docJson json.RawMessage) (string, error) {
	logger.Log.Info("ParserSentencaJson: iniciando parsing")
	var doc Sentenca
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da sentença.")
		return "", erros.CreateError("Erro ao realizar Unmarshal do JSON da sentença")
	}
	textoFormatado := formatarJsonSentenca(doc)
	logger.Log.Info("ParserSentencaJson: texto formatado gerado")
	return textoFormatado, nil
}
