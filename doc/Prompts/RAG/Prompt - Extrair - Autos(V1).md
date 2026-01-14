## Prompt - Extrair - Autos(V1)

🎯 Objetivo

Ler uma peça ou processo judicial (petição inicial, contestação, réplica, despacho, decisão, sentença, etc.) e responder somente com o JSON do tipo identificado, com dados literais e fiéis ao texto.

⚖️ Regras Gerais

Não invente, não deduza, não "complemente".

Linguagem formal jurídica.

Preencha todos os campos obrigatórios; se ausente, use "NID".

Consistência entre campos (pedidos ↔ fundamentos ↔ valores).

Identifique automaticamente o tipo de peça.

Saída = somente o JSON, sem comentários, explicações ou markdown.

🖋️ Assinatura Eletrônica

Extraia as informações do rodapé do documento:

"assinatura_data" → data e hora literal da linha "Assinado eletronicamente por".

"assinatura_por" → nome completo de quem assinou eletronicamente.

Se não houver assinatura visível, use "NID".

📌 Constantes

NID = "informação não identificada no documento".

ID_PJE: extrair o número localizado na linha do rodapé que contém "Num." antes de "- Pág.". 
O número pode ter entre 6 e 12 dígitos. 
Retorne somente os dígitos. 
Se não houver esse identificador, use "id_pje não identificado".

🗂️ Tipos de Documento (tabela)
[
  {"key":1,"description":"Petição inicial"},
  {"key":2,"description":"Contestação"},
  {"key":3,"description":"Réplica"},
  {"key":4,"description":"Despacho"},
  {"key":5,"description":"Petição"},
  {"key":6,"description":"Decisão"},
  {"key":7,"description":"Sentença"},
  {"key":8,"description":"Embargos de declaração"},
  {"key":9,"description":"Recurso de Apelação"},
  {"key":10,"description":"Contra-razões"},
  {"key":11,"description":"Procuração"},
  {"key":12,"description":"Rol de Testemunhas"},
  {"key":13,"description":"Contrato"},
  {"key":14,"description":"Laudo Pericial"},
  {"key":15,"description":"Termo de Audiência"},
  {"key":16,"description":"Parecer do Ministério Público"},
  {"key":1000,"description":"Autos Processuais"}
]
Componentes Reutilizáveis
Pessoa: {"nome":string,"cpf":string,"cnpj":string,"endereco":string}
Advogado: {"nome":string,"oab":string}
Jurisprudencia: {"sumulas":[string],"acordaos":[{"tribunal":string,"processo":string,"ementa":string,"relator":string,"data":string}]}
Deliberado: {"finalidade":string,"destinatario":string,"prazo":string}

Esquema Base (presente em todos os tipos)
{
  "tipo": {"key": number, "description": string},
  "processo": string,
  "id_pje": string,
  "assinatura_data": string,
  "assinatura_por": string
}
Campos por Tipo (adicionados ao Esquema Base)
1️⃣ Petição inicial (1)
{
  "partes": {"autor":[Pessoa],"reu":[Pessoa]},
  "pedidos": [string],
  "fatos":[string],
  "fundamentacao": [string],
  "valor_causa": string
}
2️⃣ Contestação (2)
{
  "partes": {"reu":[Pessoa],"autor":[Pessoa]},
  "preliminares": [string],
  "versao_dos_fatos":[string],
  "merito": [string],
  "pedidos": [string]
}
3️⃣ Réplica (3)
{
  "impugnacoes": [string],
  "pedidos_finais": [string]
}
4️⃣ Despacho (4)
{
  "fundamentacao": [string],
  "deliberacoes": [Deliberado]
}
5️⃣ Petição (5)
{
  "fundamentacao": [string],
  "requerimentos": [string]
}
6️⃣ Decisão (6)
{
  "fundamentacao": [string],
  "dispositivo": [string]
}
7️⃣ Sentença (7)
{
  "metadados": {
    "numero": string,
    "classe": string,
    "assunto": string,
    "juizo": string,
    "partes": {
      "autor": [string],
      "reu": [string]
    }
  },
  "questoes": [
    {
      "tipo": "string (preliminar ou mérito)",
      "tema": "string",
      "paragrafos": [string],
      "decisao": "string"
    }
  ],
  "dispositivo": {
    "paragrafos": [string]
  }
}
8️⃣ Embargos de Declaração (8)
{
  "fundamentacao": [string],
  "decisao": [string]
}

9️⃣ Recurso de Apelação (9)
{
  "fundamentos": [string],
  "pedidos": [string]
}

🔟 Contra-razões (10)
{
  "argumentos": [string],
  "requerimentos": [string]
}

1️⃣1️⃣ Procuração (11)
{
  "outorgantes": [Pessoa],
  "advogados": [Advogado],
  "poderes": [string]
}

1️⃣2️⃣ Rol de Testemunhas (12)
{
  "testemunhas": [Pessoa]
}

1️⃣3️⃣ Contrato (13)
{
  "partes": [Pessoa],
  "objeto": string,
  "clausulas": [string]
}

1️⃣4️⃣ Laudo Pericial (14)
{
  "peritos": [Pessoa],
  "quesitos": [
    {
      "numero": "string",
      "parte": "string",
      "quesito": "string",
      "resposta": "string"
    }
  ],
  "conclusoes": "string"
}

1️⃣5️⃣ Termo de Audiência (15)
{
  "data_audiencia": string,
  "tipo_audiencia": string,
  "ocorrencias": [string],
  "deliberacoes": [Deliberado]
}

1️⃣6️⃣ Parecer do Ministério Público (16)
{
  "fundamentacao": [string],
  "opiniao": string
}

1️⃣000 Autos Processuais (1000)
{
  "documentos": [string]
}

⚙️ Instruções Detalhadas ao Modelo

Identifique o tipo da peça (campo "tipo") conforme o conteúdo do texto.

Aplique o Esquema Base a todos os tipos.

Acrescente os campos específicos conforme o tipo identificado.

Transcreva fielmente o conteúdo textual dos parágrafos, fatos, fundamentos, pedidos, quesitos, etc.

Os fatos devem ser descritos com a maior riqueza de detalhes possível.

Nunca omita o dispositivo ou a conclusão.

Mantenha a data e o nome da assinatura eletrônica conforme aparecem no rodapé.

Saída = JSON válido, sem comentários, sem formatação adicional.

✅ Exemplo de saída esperada (tipo: decisão)
{
  "tipo": {
    "key": 6,
    "description": "Decisão"
  },
  "processo": "0202941-41.2024.8.06.0167",
  "id_pje": "110934355",
  "assinatura_data": "31/05/2024 18:08:32",
  "assinatura_por": "ALDENOR SOMBRA DE OLIVEIRA",
  "fundamentacao": [
    "Considerando os elementos de prova apresentados...",
    "A tutela de urgência será concedida se presentes os requisitos..."
  ],
  "dispositivo": [
    "Ante o exposto, defiro a tutela de urgência pleiteada.",
    "Intimem-se as partes."
  ]
}

