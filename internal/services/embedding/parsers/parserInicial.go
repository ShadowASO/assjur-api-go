package parsers

import (
	"encoding/json"
	"strings"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
)

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonInicial(doc PeticaoInicial) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	sb.WriteString("Natureza Jurídica: " + doc.Natureza.NomeJuridico + "; ")

	// Fatos
	sb.WriteString("Fatos: " + doc.Fatos + "; ")

	// Preliminares
	if len(doc.Preliminares) > 0 {
		sb.WriteString("Preliminares: ")
		for _, v := range doc.Preliminares {
			sb.WriteString(v + "; ")
		}
	}

	// Atos normativos
	if len(doc.AtosNormativos) > 0 {
		sb.WriteString("Atos Normativos: ")
		for _, v := range doc.AtosNormativos {
			sb.WriteString(v + "; ")
		}
	}

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos: ")
		for _, v := range doc.Pedidos {
			sb.WriteString(v + "; ")
		}
	}

	// Tutela provisória
	sb.WriteString("Tutela Provisória: " + doc.TutelaProvisoria.Detalhes + "; ")

	return sb.String()
}

// Parser para JSON do tipo Petição Inicial
func ParserInicialJson(idNatu int, docJson json.RawMessage) (string, error) {
	logger.Log.Info("ParserInicialJson iniciado")
	var doc PeticaoInicial
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Errorf("Erro ao realizar Unmarshal do JSON da inicial: ", err)
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonInicial(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
