package parsers

import (
	"encoding/json"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonApelacao(doc RecursoApelacao) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	if doc.JuizoDestinatario != "" {
		sb.WriteString("\nJuízo destinatário: " + doc.JuizoDestinatario + "; ")
	}

	if len(doc.Partes.Recorrentes) > 0 {
		sb.WriteString("\nRecorrentes: ")
		for _, p := range doc.Partes.Recorrentes {
			sb.WriteString(p.Nome + "; ")
		}
	}

	if len(doc.Partes.Recorridos) > 0 {
		sb.WriteString("\nRecorridos: ")
		for _, p := range doc.Partes.Recorridos {
			sb.WriteString(p.Nome + "; ")
		}
	}

	sb.WriteString("\nCausa de Pedir:\n" + doc.CausaDePedir + "; ")

	if len(doc.Pedidos) > 0 {
		sb.WriteString("\nPedidos: ")
		for _, p := range doc.Pedidos {
			sb.WriteString(p + "; ")
		}
	}

	if len(doc.Advogados) > 0 {
		sb.WriteString("\nAdvogados: ")
		for _, adv := range doc.Advogados {
			sb.WriteString(adv.Nome + " (OAB: " + adv.OAB + "); ")
		}
	}

	return sb.String()
}

func ParserApelacaoJson(idNatu int, docJson json.RawMessage) (string, error) {
	var doc RecursoApelacao
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON do recurso.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON do recurso")
	}
	textoFormatado := formatarJsonApelacao(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
