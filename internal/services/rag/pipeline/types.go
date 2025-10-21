package pipeline

type TipoEvento struct {
	Evento    int    `json:"evento"`
	Descricao string `json:"descricao"`
}

// Tipo para ser utilizada no protocolo RAG, quando devolver uma mensagem ao cliente
type MensagemEvento struct {
	Tipo     TipoEvento `json:"tipo"`  // Identificador do evento
	Conteudo string     `json:"texto"` // Texto de confirmação (pergunta ou afirmação)
}

type ConfirmaEvento struct {
	Tipo        TipoEvento `json:"tipo"`        // Identificador do evento
	Confirmacao string     `json:"confirmacao"` // Texto de confirmação (pergunta ou afirmação)
}
type ComplementoEvento struct {
	Tipo      TipoEvento `json:"tipo"` // Identificador do evento
	Faltantes []string   `json:"faltantes"`
}

// var verif struct {
// 		Tipo struct {
// 			Evento    int    `json:"evento"`
// 			Descricao string `json:"descricao"`
// 		} `json:"tipo"`
// 		Faltantes []string `json:"faltantes"`
// 	}

type TPartes struct {
	Autor []string `json:"autor,omitempty"`
	Reu   []string `json:"reu,omitempty"`
}

type AnaliseJuridicaIA struct {
	Tipo struct {
		Evento    int    `json:"evento"`
		Descricao string `json:"descricao"`
	} `json:"tipo"`

	Identificacao struct {
		NumeroProcesso string `json:"numero_processo"`
		Natureza       string `json:"natureza"`
	} `json:"identificacao"`

	Partes TPartes `json:"partes"`

	SinteseFatos struct {
		Autor string `json:"autor"`
		Reu   string `json:"reu"`
	} `json:"sintese_fatos"`

	PedidosAutor []string `json:"pedidos_autor"`

	DefesasReu struct {
		Preliminares       []string `json:"preliminares"`
		PrejudiciaisMerito []string `json:"prejudiciais_merito"`
		DefesaMerito       []string `json:"defesa_merito"`
		PedidosReu         []string `json:"pedidos_reu"`
	} `json:"defesas_reu"`

	QuestoesControvertidas []struct {
		Descricao         string `json:"descricao"`
		PerguntaAoUsuario string `json:"pergunta_ao_usuario"`
	} `json:"questoes_controvertidas"`

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
	ValorDaCausa        string   `json:"valor_da_causa"`
	Observacoes         []string `json:"observacoes"`

	// Campo RAG: síntese de temas jurídicos relevantes extraídos das peças processuais
	Rag []struct {
		Tema       string `json:"tema"`
		Descricao  string `json:"descricao"`
		Relevancia string `json:"relevancia"`
	} `json:"rag"`

	// Campo opcional para armazenamento dos vetores de embeddings (gerados posteriormente)
	RagEmbedding []float64 `json:"rag_embedding"`
	DataGeracao  string    `json:"data_geracao"`
}

//SENTENÇA

type MinutaSentenca struct {
	Tipo          *TipoEvento    `json:"tipo,omitempty"`
	Processo      *Processo      `json:"processo,omitempty"`
	Partes        *TPartes       `json:"partes,omitempty"`
	Relatorio     []string       `json:"relatorio,omitempty"`
	Fundamentacao *Fundamentacao `json:"fundamentacao,omitempty"`
	Dispositivo   *Dispositivo   `json:"dispositivo,omitempty"`
	Observacoes   []string       `json:"observacoes,omitempty"`
	DataGeracao   string         `json:"data_geracao"`
}

type Processo struct {
	Numero  *string `json:"numero,omitempty"`
	Classe  *string `json:"classe,omitempty"`
	Assunto *string `json:"assunto,omitempty"`
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

//*****   SENTENÇA - Extraída dos Autos do Processo.
//Esta sentença é utilizada para formar a base de conhecimentos

type SentencaAutos struct {
	Tipo *struct {
		Key         int    `json:"key"`
		Description string `json:"description"`
	} `json:"tipo"`

	Processo       string     `json:"processo"`
	IdPje          string     `json:"id_pje"`
	AssinaturaData string     `json:"assinatura_data"`
	AssinaturaPor  string     `json:"assinatura_por"`
	Metadados      *metadados `json:"metadados"`

	Questoes    []questao   `json:"questoes"`
	Dispositivo dispositivo `json:"dispositivo"`
}

type metadados struct {
	Classe  string   `json:"classe"`
	Assunto string   `json:"assunto"`
	Juizo   string   `json:"juizo"`
	Partes  *TPartes `json:"partes"`
}

type questao struct {
	Tipo       string   `json:"tipo"` // "preliminar" ou "mérito"
	Tema       string   `json:"tema"`
	Paragrafos []string `json:"paragrafos"`
	Decisao    string   `json:"decisao"`
}

type dispositivo struct {
	Paragrafos []string `json:"paragrafos"`
}
