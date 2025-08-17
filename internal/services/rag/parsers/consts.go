package parsers

type TipoDocumento struct {
	Key         int    `json:"key"`
	Description string `json:"description"`
}

type Pessoa struct {
	Nome     string `json:"nome"`
	CPF      string `json:"cpf"`
	CNPJ     string `json:"cnpj"`
	Endereco string `json:"endereco"`
}

type Advogado struct {
	Nome string `json:"nome"`
	OAB  string `json:"oab"`
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

// -----------------------
// a) Petição Inicial
type PeticaoInicial struct {
	Tipo     TipoDocumento `json:"tipo"`
	Processo string        `json:"processo"`
	IdPje    string        `json:"id_pje"`
	Natureza struct {
		NomeJuridico string `json:"nome_juridico"`
	} `json:"natureza"`
	Partes struct {
		Autor []Pessoa `json:"autor"`
		Reu   []Pessoa `json:"reu"`
	} `json:"partes"`
	Fatos            string         `json:"fatos"`
	Preliminares     []string       `json:"preliminares"`
	AtosNormativos   []string       `json:"atos_normativos"`
	Jurisprudencia   Jurisprudencia `json:"jurisprudencia"`
	Doutrina         []string       `json:"doutrina"`
	Pedidos          []string       `json:"pedidos"`
	TutelaProvisoria struct {
		Detalhes string `json:"detalhes"`
	} `json:"tutela_provisoria"`
	Provas         []string   `json:"provas"`
	RolTestemunhas []string   `json:"rol_testemunhas"`
	ValorDaCausa   string     `json:"valor_da_causa"`
	Advogados      []Advogado `json:"advogados"`
}

// -----------------------
// b) Contestação
type Contestacao struct {
	Tipo     TipoDocumento `json:"tipo"`
	Processo string        `json:"processo"`
	IdPje    string        `json:"id_pje"`
	Partes   struct {
		Autor []Pessoa `json:"autor"`
		Reu   []Pessoa `json:"reu"`
	} `json:"partes"`
	Fatos            string         `json:"fatos"`
	Preliminares     []string       `json:"preliminares"`
	AtosNormativos   []string       `json:"atos_normativos"`
	Jurisprudencia   Jurisprudencia `json:"jurisprudencia"`
	Doutrina         []string       `json:"doutrina"`
	Pedidos          []string       `json:"pedidos"`
	TutelaProvisoria struct {
		Detalhes string `json:"detalhes"`
	} `json:"tutela_provisoria"`
	QuestoesControvertidas []string   `json:"questoes_controvertidas"`
	Provas                 []string   `json:"provas"`
	RolTestemunhas         []string   `json:"rol_testemunhas"`
	Advogados              []Advogado `json:"advogados"`
}

// -----------------------
// c) Réplica
type Replica struct {
	Tipo                   TipoDocumento `json:"tipo"`
	Processo               string        `json:"processo"`
	IdPje                  string        `json:"id_pje"`
	PartesPeticionantes    []Pessoa      `json:"partes_peticionantes"`
	Fatos                  string        `json:"fatos"`
	QuestoesControvertidas []string      `json:"questoes_controvertidas"`
	Pedidos                []string      `json:"pedidos"`
	Provas                 []string      `json:"provas"`
	RolTestemunhas         []string      `json:"rol_testemunhas"`
	Advogados              []Advogado    `json:"advogados"`
}

// -----------------------
// d) Petição diversa
type PeticaoDiversa struct {
	Tipo                TipoDocumento `json:"tipo"`
	Processo            string        `json:"processo"`
	IdPje               string        `json:"id_pje"`
	PartesPeticionantes []Pessoa      `json:"partes_peticionantes"`
	CausaDePedir        string        `json:"causaDePedir"`
	Pedidos             []string      `json:"pedidos"`
	Advogados           []Advogado    `json:"advogados"`
}

// -----------------------
// e) Despacho (inicial e ordinatório)
type Despacho struct {
	Tipo       TipoDocumento `json:"tipo"`
	Processo   string        `json:"processo"`
	IdPje      string        `json:"id_pje"`
	Conteudo   []string      `json:"conteudo"`
	Deliberado []struct {
		Finalidade   string `json:"finalidade"`
		Destinatario string `json:"destinatario"`
		Prazo        string `json:"prazo"`
	} `json:"deliberado"`
	Juiz struct {
		Nome string `json:"nome"`
	} `json:"juiz"`
}

// -----------------------
// f) Decisão interlocutória / Tutela provisória
type DecisaoInterlocutoria struct {
	Tipo       TipoDocumento `json:"tipo"`
	Processo   string        `json:"processo"`
	IdPje      string        `json:"id_pje"`
	Conteudo   []string      `json:"conteudo"`
	Deliberado []struct {
		Finalidade   string `json:"finalidade"`
		Destinatario string `json:"destinatario"`
		Prazo        string `json:"prazo"`
	} `json:"deliberado"`
	Juiz struct {
		Nome string `json:"nome"`
	} `json:"juiz"`
}

// -----------------------
// g) Sentença
type Sentenca struct {
	Tipo         TipoDocumento `json:"tipo"`
	Processo     string        `json:"processo"`
	IdPje        string        `json:"id_pje"`
	Preliminares []struct {
		Assunto string `json:"assunto"`
		Decisao string `json:"decisao"`
	} `json:"preliminares"`
	Fundamentos []struct {
		Texto  string   `json:"texto"`
		Provas []string `json:"provas"`
	} `json:"fundamentos"`
	Conclusao []struct {
		Resultado    string `json:"resultado"`
		Destinatario string `json:"destinatario"`
		Prazo        string `json:"prazo"`
		Decisao      string `json:"decisao"`
	} `json:"conclusao"`
	Juiz struct {
		Nome string `json:"nome"`
	} `json:"juiz"`
}

// -----------------------
// h) Embargos de declaração
type EmbargosDeclaracao struct {
	Tipo     TipoDocumento `json:"tipo"`
	Processo string        `json:"processo"`
	IdPje    string        `json:"id_pje"`
	Partes   struct {
		Recorrentes []Pessoa `json:"recorrentes"`
		Recorridos  []Pessoa `json:"recorridos"`
	} `json:"partes"`
	JuizoDestinatario string     `json:"juizoDestinatario"`
	CausaDePedir      string     `json:"causaDePedir"`
	Pedidos           []string   `json:"pedidos"`
	Advogados         []Advogado `json:"advogados"`
}

// -----------------------
// i) Recurso de Apelação
type RecursoApelacao struct {
	Tipo     TipoDocumento `json:"tipo"`
	Processo string        `json:"processo"`
	IdPje    string        `json:"id_pje"`
	Partes   struct {
		Recorrentes []Pessoa `json:"recorrentes"`
		Recorridos  []Pessoa `json:"recorridos"`
	} `json:"partes"`
	JuizoDestinatario string     `json:"juizoDestinatario"`
	CausaDePedir      string     `json:"causaDePedir"`
	Pedidos           []string   `json:"pedidos"`
	Advogados         []Advogado `json:"advogados"`
}

// -----------------------
// j) Procuração
type Procuracao struct {
	Tipo        TipoDocumento `json:"tipo"`
	Processo    string        `json:"processo"`
	IdPje       string        `json:"id_pje"`
	Outorgantes []Pessoa      `json:"outorgantes"`
	Advogados   []Advogado    `json:"advogados"`
	Poderes     string        `json:"poderes"`
}

// -----------------------
// k) Rol de testemunhas
type RolTestemunhas struct {
	Tipo        TipoDocumento `json:"tipo"`
	Processo    string        `json:"processo"`
	IdPje       string        `json:"id_pje"`
	Partes      []Pessoa      `json:"partes"`
	Testemunhas []Pessoa      `json:"testemunhas"`
	Advogados   []Advogado    `json:"advogados"`
}

// -----------------------
// l) Laudo Pericial
type LaudoPericial struct {
	Tipo       TipoDocumento `json:"tipo"`
	Processo   string        `json:"processo"`
	IdPje      string        `json:"id_pje"`
	Peritos    []Pessoa      `json:"peritos"`
	Conclusoes string        `json:"conclusoes"`
}

// TipoDocumento e structs auxiliares assumidos já definidos:
// type TipoDocumento struct { Key int `json:"key"` Description string `json:"description"` }
// type Pessoa struct { Nome string `json:"nome"` Qualidade string `json:"qualidade"` }
type Manifestacao struct {
	Nome         string `json:"nome"`
	Manifestacao string `json:"manifestacao"`
}

type TermoAudiencia struct {
	Tipo          TipoDocumento  `json:"tipo"`
	Processo      string         `json:"processo"`
	IdPje         string         `json:"id_pje"`
	Local         string         `json:"local"`
	Data          string         `json:"data"`
	Hora          string         `json:"hora"`
	Presentes     []Pessoa       `json:"presentes"`
	Descricao     string         `json:"descricao"`
	Manifestacoes []Manifestacao `json:"manifestacoes"`
}
