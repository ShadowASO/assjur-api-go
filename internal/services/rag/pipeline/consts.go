package pipeline

type AnaliseProcesso struct {
	Tipo struct {
		Codigo    int    `json:"codigo"`
		Descricao string `json:"descricao"`
	} `json:"tipo"`

	Identificacao struct {
		NumeroProcesso string `json:"numero_processo"`
		Natureza       string `json:"natureza"`
	} `json:"identificacao"`

	Partes struct {
		Autor struct {
			Nome         string `json:"nome"`
			Qualificacao string `json:"qualificacao"`
			Endereco     string `json:"endereco"`
		} `json:"autor"`
		Reu struct {
			Nome     string `json:"nome"`
			CNPJ     string `json:"cnpj"`
			Endereco string `json:"endereco"`
		} `json:"reu"`
	} `json:"partes"`

	SinteseFatos struct {
		Autor string `json:"autor"`
		Reu   string `json:"reu"`
	} `json:"sintese_fatos"`

	PedidosAutor []string `json:"pedidos_autor"`

	DefesasReu struct {
		Preliminares        []string `json:"preliminares"`
		PrejudiciaisMerito  []string `json:"prejudiciais_merito"`
		DefesaMerito        []string `json:"defesa_merito"`
		PedidosSubsidiarios []string `json:"pedidos_subsidiarios"`
	} `json:"defesas_reu"`

	QuestoesControvertidas []string `json:"questoes_controvertidas"`

	Provas struct {
		Autor []string `json:"autor"`
		Reu   []string `json:"reu"`
	} `json:"provas"`

	FundamentacaoJuridica struct {
		Autor          []string `json:"autor"`
		Reu            []string `json:"reu"`
		Jurisprudencia []struct {
			Tribunal string `json:"tribunal"`
			Processo string `json:"processo"`
			Tema     string `json:"tema"`
			Ementa   string `json:"ementa"`
		} `json:"jurisprudencia"`
	} `json:"fundamentacao_juridica"`

	DecisoesInterlocutorias []struct {
		IdDecisao     string `json:"id_decisao"`
		Conteudo      string `json:"conteudo"`
		Magistrado    string `json:"magistrado"`
		Fundamentacao string `json:"fundamentacao"`
	} `json:"decisoes_interlocutorias"`

	AndamentoProcessual []string `json:"andamento_processual"`

	ValorDaCausa string `json:"valor_da_causa"`

	Observacoes []string `json:"observacoes"`
}

//SENTENÇA

type SentencaRAG struct {
	Tipo          *Tipo          `json:"tipo,omitempty"`
	Processo      *Processo      `json:"processo,omitempty"`
	Partes        *Partes        `json:"partes,omitempty"`
	Relatorio     []string       `json:"relatorio,omitempty"`
	Fundamentacao *Fundamentacao `json:"fundamentacao,omitempty"`
	Dispositivo   *Dispositivo   `json:"dispositivo,omitempty"`
	Observacoes   []string       `json:"observacoes,omitempty"`
	Assinatura    *Assinatura    `json:"assinatura,omitempty"`
}

type Tipo struct {
	Codigo    *int    `json:"codigo,omitempty"`
	Descricao *string `json:"descricao,omitempty"`
}

type Processo struct {
	Numero  *string `json:"numero,omitempty"`
	Classe  *string `json:"classe,omitempty"`
	Assunto *string `json:"assunto,omitempty"`
}

type Partes struct {
	Autor []string `json:"autor,omitempty"`
	Reu   []string `json:"reu,omitempty"`
}

type Fundamentacao struct {
	Preliminares   []string        `json:"preliminares,omitempty"`
	Merito         []string        `json:"merito,omitempty"`
	Doutrina       []string        `json:"doutrina,omitempty"`
	Jurisprudencia *Jurisprudencia `json:"jurisprudencia,omitempty"`
}

type Jurisprudencia struct {
	Sumulas  []string  `json:"sumulas,omitempty"`
	Acordaos []Acordao `json:"acordaos,omitempty"`
}

type Acordao struct {
	Tribunal *string `json:"tribunal,omitempty"`
	Processo *string `json:"processo,omitempty"`
	Ementa   *string `json:"ementa,omitempty"`
	Relator  *string `json:"relator,omitempty"`
	Data     *string `json:"data,omitempty"`
}

type Dispositivo struct {
	Decisao     *string  `json:"decisao,omitempty"`
	Condenacoes []string `json:"condenacoes,omitempty"`
	Honorarios  *string  `json:"honorarios,omitempty"`
	Custas      *string  `json:"custas,omitempty"`
}

type Assinatura struct {
	Juiz  *string `json:"juiz,omitempty"`
	Cargo *string `json:"cargo,omitempty"`
}
