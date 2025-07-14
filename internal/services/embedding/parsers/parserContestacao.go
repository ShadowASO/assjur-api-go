package parsers

import (
	"encoding/json"

	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

type Contestacao struct {
	Tipo                   Tipo              `json:"tipo"`
	Processo               string            `json:"processo"`
	IdPje                  string            `json:"id_pje"`
	Partes                 PartesContestacao `json:"partes"`
	Fatos                  string            `json:"fatos"`
	Preliminares           []string          `json:"preliminares"`
	AtosNormativos         []string          `json:"atos_normativos"`
	Jurisprudencia         Jurisprudencia    `json:"jurisprudencia"`
	Doutrina               []string          `json:"doutrina"`
	Pedidos                []string          `json:"pedidos"`
	TutelaProvisoria       TutelaProvisoria  `json:"tutela_provisoria"`
	QuestoesControvertidas []string          `json:"questoes_controvertidas"`
	Provas                 []string          `json:"provas"`
	RolTestemunhas         []string          `json:"rol_testemunhas"`
	Advogados              []Advogado        `json:"advogados"`
}

// Função que limpa dados sensíveis e monta o texto para embedding
func formatarJsonContestacao(doc Contestacao) string {
	var sb strings.Builder

	//Natureza do documento: Contestação
	sb.WriteString(doc.Tipo.Description + ": ")

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

	// Questões controvertidas
	if len(doc.QuestoesControvertidas) > 0 {
		sb.WriteString("Questões Controvertidas: ")
		for _, v := range doc.QuestoesControvertidas {
			sb.WriteString(v + "; ")
		}

	}

	return sb.String()
}

func ParserContestacaoJson(idNatu int, docJson json.RawMessage) (string, error) {
	//logger.Log.Info("Entrei")

	var doc Contestacao
	err := json.Unmarshal(docJson, &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}
	textoFormatado := formatarJsonContestacao(doc)
	logger.Log.Info(textoFormatado)
	return textoFormatado, nil
}
