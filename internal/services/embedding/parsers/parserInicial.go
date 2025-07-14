package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Estruturas para receber o JSON (com os campos necessários)
type Inicial struct {
	Tipo             Tipo             `json:"tipo"`
	Processo         string           `json:"processo"`
	IdPje            string           `json:"id_pje"`
	Natureza         Natureza         `json:"natureza"`
	Partes           Partes           `json:"partes"`
	Fatos            string           `json:"fatos"`
	Preliminares     []string         `json:"preliminares"`
	AtosNormativos   []string         `json:"atos_normativos"`
	Jurisprudencia   Jurisprudencia   `json:"jurisprudencia"`
	Doutrina         []string         `json:"doutrina"`
	Pedidos          []string         `json:"pedidos"`
	TutelaProvisoria TutelaProvisoria `json:"tutela_provisoria"`
	Provas           []string         `json:"provas"`
	RolTestemunhas   []string         `json:"rol_testemunhas"`
	ValorDaCausa     string           `json:"valor_da_causa"`
	Advogados        []Advogado       `json:"advogados"`
}

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonInicial(doc Inicial) string {
	var sb strings.Builder

	sb.WriteString(doc.Tipo.Description + ": ")

	//sb.WriteString("Tipo: " + doc.Tipo.Description + "\n")
	//sb.WriteString("Processo: " + doc.Processo + "\n")
	sb.WriteString("Natureza Jurídica: " + doc.Natureza.NomeJuridico + "; ")

	// Fatos
	sb.WriteString("Fatos: " + doc.Fatos + "; ")

	// Preliminares
	if len(doc.Preliminares) > 0 {
		sb.WriteString("Preliminares: ")
		for _, v := range doc.Preliminares {
			sb.WriteString(v + "; ")
		}
		//sb.WriteString("\n")
	}

	// Atos normativos
	if len(doc.AtosNormativos) > 0 {
		sb.WriteString("Atos Normativos: ")
		for _, v := range doc.AtosNormativos {
			sb.WriteString(v + "; ")
		}
		//sb.WriteString("\n")
	}

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos: ")
		for _, v := range doc.Pedidos {
			sb.WriteString(v + "; ")
		}
		//sb.WriteString("\n")
	}

	// Tutela provisória
	sb.WriteString("Tutela Provisória:\n" + doc.TutelaProvisoria.Detalhes + "; ")

	return sb.String()
}

func ParserInicialJson(idNatu int, docJson json.RawMessage) (string, error) {
	logger.Log.Info("Entrei")
	var doc Inicial
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonInicial(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
