package parsers

type Tipo struct {
	Key         int    `json:"key"`
	Description string `json:"description"`
}

type Conteudo []string

type Deliberado struct {
	Finalidade   string `json:"finalidade"`
	Destinatario string `json:"destinatario"`
	Prazo        string `json:"prazo"`
}

type Juiz struct {
	Nome string `json:"nome"`
}

type Advogado struct {
	Nome string `json:"nome"`
	OAB  string `json:"OAB"`
}
type Parte struct {
	Nome string `json:"nome"`
	CPF  string `json:"CPF"`
	CNPJ string `json:"CNPJ"`
	End  string `json:"end"`
}

type Partes struct {
	Requerentes []Parte `json:"requerentes"`
	Requeridos  []Parte `json:"requeridos"`
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

type Natureza struct {
	NomeJuridico string `json:"nome_juridico"`
}

type TutelaProvisoria struct {
	Detalhes string `json:"detalhes"`
}

type PartesContestacao struct {
	Contestantes []Parte `json:"contestantes"`
	Contestados  []Parte `json:"contestados"`
}
