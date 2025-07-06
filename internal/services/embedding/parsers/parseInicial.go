package parsers

import (
	"encoding/json"
	"fmt"
	"ocrserver/internal/utils/erros"
	"ocrserver/internal/utils/logger"
	"strings"
)

// Estruturas para receber o JSON (com os campos necessários)
type Documento struct {
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

type Tipo struct {
	Key         int    `json:"key"`
	Description string `json:"description"`
}

type Natureza struct {
	NomeJuridico string `json:"nome_juridico"`
}

type Partes struct {
	Requerentes []Parte `json:"requerentes"`
	Requeridos  []Parte `json:"requeridos"`
}

type Parte struct {
	Nome string `json:"nome"`
	CPF  string `json:"CPF"`
	CNPJ string `json:"CNPJ"`
	End  string `json:"end"`
}

type Jurisprudencia struct {
	Sumulas  []string  `json:"sumulas"`
	Acordaos []Acordao `json:"acordaos"`
}

type Acordao struct {
	Tribunal string `json:"tribunal"`
	Processo string `json:"processo"`
	Ementa   string `json:"ementa"`
	Relator  string `json:"relator"`
	Data     string `json:"data"`
}

type TutelaProvisoria struct {
	Detalhes string `json:"detalhes"`
}

type Advogado struct {
	Nome string `json:"nome"`
	OAB  string `json:"OAB"`
}

// Função que limpa dados sensíveis e monta o texto para embedding
func MontarTextoParaEmbedding(doc Documento) string {
	var sb strings.Builder

	sb.WriteString("Tipo: " + doc.Tipo.Description + "\n")
	sb.WriteString("Processo: " + doc.Processo + "\n")
	sb.WriteString("Natureza Jurídica: " + doc.Natureza.NomeJuridico + "\n")

	// Partes (sem CPF/CNPJ)
	// sb.WriteString("Requerentes:\n")
	// for _, p := range doc.Partes.Requerentes {
	// 	sb.WriteString("- Nome: " + p.Nome + "\n")
	// 	sb.WriteString("  Endereço: " + p.End + "\n")
	// }
	// sb.WriteString("\nRequeridos:\n")
	// for _, p := range doc.Partes.Requeridos {
	// 	sb.WriteString("- Nome: " + p.Nome + "\n")
	// 	sb.WriteString("  Endereço: " + p.End + "\n")
	// }
	// sb.WriteString("\n")

	// Fatos
	sb.WriteString("Fatos:\n" + doc.Fatos + "\n")

	// Preliminares
	if len(doc.Preliminares) > 0 {
		sb.WriteString("Preliminares:\n")
		for _, v := range doc.Preliminares {
			sb.WriteString("- " + v + "\n")
		}
		sb.WriteString("\n")
	}

	// Atos normativos
	if len(doc.AtosNormativos) > 0 {
		sb.WriteString("Atos Normativos:\n")
		for _, v := range doc.AtosNormativos {
			sb.WriteString("- " + v + "\n")
		}
		sb.WriteString("\n")
	}

	// Jurisprudência - ementas dos acórdãos
	if len(doc.Jurisprudencia.Acordaos) > 0 {
		sb.WriteString("Jurisprudência:\n")
		for i, a := range doc.Jurisprudencia.Acordaos {
			sb.WriteString(fmt.Sprintf("Acórdão %d:\n", i+1))
			sb.WriteString("Tribunal: " + a.Tribunal + "\n")
			sb.WriteString("Ementa: " + a.Ementa + "\n\n")
		}
	}

	// Doutrina (se houver)
	if len(doc.Doutrina) > 0 {
		sb.WriteString("Doutrina:\n")
		for _, v := range doc.Doutrina {
			sb.WriteString("- " + v + "\n")
		}
		sb.WriteString("\n")
	}

	// Pedidos
	if len(doc.Pedidos) > 0 {
		sb.WriteString("Pedidos:\n")
		for _, v := range doc.Pedidos {
			sb.WriteString("- " + v + "\n")
		}
		sb.WriteString("\n")
	}

	// Tutela provisória
	sb.WriteString("Tutela Provisória:\n" + doc.TutelaProvisoria.Detalhes + "\n\n")

	// Provas
	if len(doc.Provas) > 0 {
		sb.WriteString("Provas:\n")
		for _, v := range doc.Provas {
			sb.WriteString("- " + v + "\n")
		}
		sb.WriteString("\n")
	}

	// Rol testemunhas
	// if len(doc.RolTestemunhas) > 0 {
	// 	sb.WriteString("Rol de Testemunhas:\n")
	// 	for _, v := range doc.RolTestemunhas {
	// 		sb.WriteString("- " + v + "\n")
	// 	}
	// 	sb.WriteString("\n")
	// }

	// Valor da causa
	if doc.ValorDaCausa != "" {
		sb.WriteString("Valor da Causa: " + doc.ValorDaCausa + "\n\n")
	}

	// Advogados
	// if len(doc.Advogados) > 0 {
	// 	sb.WriteString("Advogados:\n")
	// 	for _, adv := range doc.Advogados {
	// 		sb.WriteString("- Nome: " + adv.Nome + ", OAB: " + adv.OAB + "\n")
	// 	}
	// 	sb.WriteString("\n")
	// }

	return sb.String()
}

func FormataInicialToEmbedding(idNatu int, docJson string) (string, error) {

	var doc Documento
	err := json.Unmarshal([]byte(docJson), &doc)
	if err != nil {
		logger.Log.Error("Erro ao realizar Unmarshal do JSON da inicial.")
		return "", erros.CreateError("Erro ao realizar Unmarshal de JSON da inicial")
	}

	textoFormatado := MontarTextoParaEmbedding(doc)
	//fmt.Println(textoEmbedding)
	return textoFormatado, nil
}
