package parsers

import (
	"encoding/json"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonEmbargosDeclaracao(doc EmbargosDeclaracao) string {
	var sb strings.Builder

	// Natureza do documento
	sb.WriteString(doc.Tipo.Description + ": ")

	// Juízo destinatário
	if doc.JuizoDestinatario != "" {
		sb.WriteString("\nJuízo destinatário: " + doc.JuizoDestinatario + "; ")
	}

	// Partes - recorrentes
	if len(doc.Partes.Recorrentes) > 0 {
		sb.WriteString("\nRecorrentes: ")
		for _, p := range doc.Partes.Recorrentes {
			sb.WriteString(p.Nome + "; ")
		}
	}

	// Partes - recorridos
	if len(doc.Partes.Recorridos) > 0 {
		sb.WriteString("\nRecorridos: ")
		for _, p := range doc.Partes.Recorridos {
			sb.WriteString(p.Nome + "; ")
		}
	}

	// Causa de pedir
	sb.WriteString("\nCausa de Pedir:\n" + doc.CausaDePedir + "; ")

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("\nPedidos: ")
		for _, p := range doc.Pedidos {
			sb.WriteString(p + "; ")
		}
	}

	// Advogados
	if len(doc.Advogados) > 0 {
		sb.WriteString("\nAdvogados: ")
		for _, adv := range doc.Advogados {
			sb.WriteString(adv.Nome + " (OAB: " + adv.OAB + "); ")
		}
	}

	return sb.String()
}

func ParserEmbargosDeclaracaoJson(idNatu int, docJson json.RawMessage) (string, error) {

	var doc EmbargosDeclaracao
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonEmbargosDeclaracao(doc)
	//logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
