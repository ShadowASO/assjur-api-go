🧠 Prompt Completo para Extração de Dados Jurídicos em JSON
⚖️ OBJETIVO GERAL
Você receberá um documento jurídico (ex.: petição inicial, contestação, decisão etc.) e deverá extrair as informações relevantes de forma literal e fiel ao conteúdo, preenchendo o JSON adequado de acordo com o tipo de peça identificada.

🚨 REGRAS GERAIS
Jamais invente, deduza ou complete informações ausentes.

Use linguagem formal e jurídica.

Preencha todos os campos obrigatórios. Caso a informação não conste no documento, escreva: "informação não identificada no documento".

Mantenha consistência entre os campos (ex: pedidos, valores, fundamentos, jurisprudência).

Não inclua comentários fora do JSON.

Não use blocos de código, como ```json.

Responda somente com o conteúdo do JSON gerado.

🔍 SOBRE O CAMPO id_pje
Trata-se de um número de exatamente 9 dígitos, que aparece no rodapé próximo a: Num. ######### - Pág.

Extraia somente os 9 dígitos numéricos.

Exemplo: Num. 124984094 - Pág. 2 → "124984094"

Caso não apareça nesse formato, use: "id_pje não identificado".

✅ CHECKLIST FINAL
 Todos os campos obrigatórios preenchidos?

 Nenhuma informação presumida?

 Termos jurídicos mantidos com exatidão?

 Valores, datas e fundamentos incluídos conforme aparecem no texto?

 Nenhuma omissão de jurisprudência, doutrina ou normativos citados?



## 🧩 TABELA DE TIPOS DE DOCUMENTOS
[
  { "key": 1, "description": "Petição inicial" },
  { "key": 2, "description": "Contestação" },
  { "key": 3, "description": "Réplica" },
  { "key": 4, "description": "Despacho" }, 
  { "key": 5, "description": "Petição" },
  { "key": 6, "description": "Decisão" },
  { "key": 7, "description": "Sentença" },
  { "key": 8, "description": "Embargos de declaração" },
  { "key": 9, "description": "Recurso de Apelação" },
  { "key": 10, "description": "Contra-razões" },
  { "key": 11, "description": "Procuração" },
  { "key": 12, "description": "Rol de Testemunhas" },
  { "key": 13, "description": "Contrato" },
  { "key": 14, "description": "Laudo Pericial" },
  { "key": 15, "description": "Termo de Audiência" },
  { "key": 16, "description": "Parecer do Ministério Público" },
  { "key": 1000, "description": "Autos Processuais" }
]


## 📦 MODELOS JSON POR TIPO DE DOCUMENTO

### a) Petição Inicial
{
  "tipo": { "key": 1, "description": "Petição inicial" },
  "processo": "string",
  "id_pje": "string",
  "natureza": {
    "nome_juridico": "string"
  },
  "partes": {
    "autor": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ],
    "reu": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ]
  },
  "fatos": "string",
  "preliminares": [
    "string"
  ],
  "atos_normativos": [
    "string"
  ],
  "jurisprudencia": {
    "sumulas": [ "string" ],
    "acordaos": [
      {
        "tribunal": "string",
        "processo": "string",
        "ementa": "string",
        "relator": "string",
        "data": "string"
      }
    ]
  },
  "doutrina": [ "string" ],
  "pedidos": [
    "string"
  ],
  "tutela_provisoria": {
    "detalhes": "string"
  },
  "provas": [
    "string"
  ],
  "rol_testemunhas": [ "string" ],
  "valor_da_causa": "string",
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}

### b) Contestação

{
  "tipo": { "key": 2, "description": "Contestação" },
  "processo": "string",
  "id_pje": "string",
  "partes": {
    "autor": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ],
    "reu": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ]
  },
  "fatos": "string",
  "preliminares": [
    "string"
  ],
  "atos_normativos": [ "string" ],
  "jurisprudencia": {
    "sumulas": [ ],
    "acordaos": [ ]
  },
  "doutrina": [ ],
  "pedidos": [ "string" ],
  "tutela_provisoria": {
    "detalhes": "string"
  },
  "questoes_controvertidas": [ "string" ],
  "provas": [ ],
  "rol_testemunhas": [ ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}


### c) Réplica

{
  "tipo": { "key": 3, "description": "Réplica" },
  "processo": "string",
  "id_pje": "string",
  "partes_peticionantes": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "fatos": "string",
  "questoes_controvertidas": [ "string" ],
  "pedidos": [ "string" ],
  "provas": [ "string" ],
  "rol_testemunhas": [ "string" ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}


### d) Petição

{
  "tipo": { "key": 5, "description": "Petição" },
  "processo": "string",
  "id_pje": "string",
  "partes_peticionantes": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "causaDePedir": "string",
  "pedidos": [ "string" ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}


### e) Despacho

{
  "tipo": { "key": 4, "description": "Despacho" },
  "processo": "string",
  "id_pje": "string",
  "conteudo": [ "string" ],
  "deliberado": [
    {
      "finalidade": "string",
      "destinatario": "string",
      "prazo": "string"
    }
  ],
  "juiz": {
    "nome": "string"
  }
}

### f) Decisão
{
  "tipo": { "key": 6, "description": "Decisão" },
  "processo": "string",
  "id_pje": "string",
  "conteudo": [ "string" ],
  "deliberado": [
    {
      "finalidade": "string",
      "destinatario": "string",
      "prazo": "string"
    }
  ],
  "juiz": {
    "nome": "string"
  }
}

### h) Sentença

{
  "tipo": { "key": 7, "description": "Sentença" },
  "processo": "string",
  "id_pje": "string",
  "preliminares": [
    {
      "assunto": "string",
      "decisao": "string"
    }
  ],
  "fundamentos": [
    {
      "texto": "string",
      "provas": [ "string" ]
    }
  ],
  "conclusao": [
    {
      "resultado": "string",
      "destinatario": "string",
      "prazo": "string",
      "decisao": "string"
    }
  ],
  "juiz": {
    "nome": "string"
  }
}

### i) embargos de declaração

{
  "tipo": { "key": 8, "description": "Embargos de declaração" },
  "processo": "string",
  "id_pje": "string",
  "partes": {
    "recorrentes": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ],
    "recorridos": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ]
  },
  "juizoDestinatario": "string",
  "causaDePedir": "string",
  "pedidos": [ "string" ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}


### i) recurso de apelação

{
  "tipo": { "key": 9, "description": "Recurso de Apelação" },
  "processo": "string",
  "id_pje": "string",
  "partes": {
    "recorrentes": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ],
    "recorridos": [
      {
        "nome": "string",
        "cpf": "string",
        "cnpj": "string",
        "endereco": "string"
      }
    ]
  },
  "juizoDestinatario": "string",
  "causaDePedir": "string",
  "pedidos": [ "string" ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}

### j) Procuração

{
  "tipo": { "key": 11, "description": "Procuração" },
  "processo": "string",
  "id_pje": "string",
  "outorgantes": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ],
  "poderes": "string"
}


### j) Rol de testemunhas

{
  "tipo": { "key": 12, "description": "Rol de Testemunhas" },
  "processo": "string",
  "id_pje": "string",
  "partes": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "testemunhas": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "advogados": [
    {
      "nome": "string",
      "oab": "string"
    }
  ]
}


### j) laudo pericial

{
  "tipo": { "key": 14, "description": "Laudo Pericial" },
  "processo": "string",
  "id_pje": "string",
  "peritos": [
    {
      "nome": "string",
      "cpf": "string",
      "cnpj": "string",
      "endereco": "string"
    }
  ],
  "conclusoes": "string"
}

### l) termo de audiência

{
  "tipo": { "key": 15, "description": "Termo de audiência" },
  "processo": "string",
  "id_pje": "string",
  "local": "string",
  "data": "string",
  "hora": "string",
  "presentes": [
    {
      "nome": "string",
      "qualidade": "juiz, requerente, requerido, advogado, conciliador, acadêmico, estudante etc"
    }
  ],
  "descricao": "Após o apregoamento das partes, o senhor Conciliador verificou a presença das partes acima citadas e considerou aberto o ato audiencial. Observou que há contestação às fls.183/200 dos presentes autos.",
  "manifestacoes": [
    {
      "nome": "string",
      "manifestacao": "string"
    }
  ]
}

Se algum campo não for encontrado no documento, use "informação não identificada no documento" como valor.

