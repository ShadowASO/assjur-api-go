
# 🧠 Prompt Completo para Extração de Dados Jurídicos em JSON

## ⚖️ OBJETIVO GERAL
Você receberá um documento jurídico (ex.: petição inicial, contestação, decisão etc.) e deverá extrair as informações relevantes **de forma literal e fiel ao conteúdo**, preenchendo o JSON adequado de acordo com o tipo de peça identificada.

## 🚨 REGRAS GERAIS
1. **Jamais invente, deduza ou complete informações ausentes**.
2. Use **linguagem formal e jurídica**.
3. **Preencha todos os campos obrigatórios**. Caso a informação não conste no documento, escreva: `"informação não identificada no documento"`.
4. Mantenha **consistência entre os campos** (ex: pedidos, valores, fundamentos, jurisprudência).
5. **Não inclua comentários fora do JSON**.
6. **Não use blocos de código**, como \`\`\`json.
7. **Responda somente com o conteúdo do JSON gerado**.

## 🔍 SOBRE O CAMPO `id_pje`
- Trata-se de um número de **exatamente 9 dígitos**, que aparece no rodapé próximo a: `Num. ######### - Pág.`  
- Extraia **somente os 9 dígitos numéricos**.
- Exemplo: `Num. 124984094 - Pág. 2` → `"124984094"`
- Caso não apareça nesse formato, use: `"id_pje não identificado"`.

## ✅ CHECKLIST FINAL
- [ ] Todos os campos obrigatórios preenchidos?
- [ ] Nenhuma informação presumida?
- [ ] Termos jurídicos mantidos com exatidão?
- [ ] Valores, datas e fundamentos incluídos conforme aparecem no texto?
- [ ] Nenhuma omissão de jurisprudência, doutrina ou normativos citados?

## 🧩 TABELA DE TIPOS DE DOCUMENTOS
```json
{ "key": 1, "description": "Petição inicial" }
{ "key": 2, "description": "Contestação" }
{ "key": 3, "description": "Réplica" }
{ "key": 4, "description": "Despacho inicial" }
{ "key": 5, "description": "Despacho ordinatório" }
{ "key": 6, "description": "Petição diversa" }
{ "key": 7, "description": "Decisão interlocutória" }
{ "key": 8, "description": "Sentença" }
{ "key": 9, "description": "Embargos de declaração" }
{ "key": 10, "description": "Contra-razões" }
{ "key": 11, "description": "Recurso de Apelação" }
{ "key": 12, "description": "Procuração" }
{ "key": 13, "description": "Rol de Testemunhas" }
{ "key": 14, "description": "Contrato" }
{ "key": 15, "description": "Laudo Pericial" }
{ "key": 1000, "description": "Autos Processuais" }
```

## 📦 MODELOS JSON POR TIPO DE DOCUMENTO

### a) Petição Inicial
{
  "tipo": { "key": 1, "description": "Petição inicial" },
  "processo": "Extrair o número de processo",
  "id_pje": "Extrair conforme regra explicada",
  "natureza": {
    "nome_juridico": "Denominação dada à ação pelo autor"
  },
  "partes": {
    "requerentes": [
      {
        "nome": "Nome completo do requerente",
        "CPF": "Número do CPF (se aplicável)",
        "CNPJ": "Número do CNPJ (se aplicável)",
        "end": "Endereço completo do requerente"
      }
    ],
    "requeridos": [
      {
        "nome": "Nome completo do requerido",
        "CPF": "Número do CPF (se aplicável)",
        "CNPJ": "Número do CNPJ (se aplicável)",
        "end": "Endereço completo do requerido"
      }
    ]
  },
  "fatos": "Descrição completa dos fatos, com datas, valores, números de contrato, etc.",
  "preliminares": [
    "Gratuidade, inversão do ônus da prova, prescrição, decadência etc."
  ],
  "atos_normativos": [
    "Citar artigos legais, constitucionais ou infralegais mencionados"
  ],
  "jurisprudencia": {
    "sumulas": [],
    "acordaos": [
      {
        "tribunal": "Nome do tribunal",
        "processo": "Número do processo",
        "ementa": "Ementa citada",
        "relator": "Nome do relator (com título)",
        "data": "Data de publicação"
      }
    ]
  },
  "doutrina": [],
  "pedidos": [
    "Pedidos formulados, com valores e fundamentos se possível"
  ],
  "tutela_provisoria": {
    "detalhes": "Descrição do pedido de tutela provisória, se houver"
  },
  "provas": [
    "Provas documentais, testemunhais, periciais etc."
  ],
  "rol_testemunhas": [],
  "valor_da_causa": "Valor total, sem símbolo R$",
  "advogados": [
    {
      "nome": "Nome do advogado",
      "OAB": "Número de registro (ex: OAB/SP 123456)"
    }
  ]
}
### b) Contestação
{
  "tipo": { "key": 2, "description": "Contestação" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "partes": {
    "contestantes": [
      {
        "nome": "Nome completo do contestante",
        "CPF": "Se aplicável",
        "CNPJ": "Se aplicável",
        "end": "Endereço"
      }
    ],
    "contestados": [
      {
        "nome": "Parte autora (contestada)"
      }
    ]
  },
  "fatos": "Versão dos fatos, com datas, valores, eventos citados",
  "preliminares": [
    "Prescrição, ilegitimidade, incompetência etc."
  ],
  "atos_normativos": [],
  "jurisprudencia": {
    "sumulas": [],
    "acordaos": []
  },
  "doutrina": [],
  "pedidos": [
    "Pedidos da defesa"
  ],
  "tutela_provisoria": {
    "detalhes": "Se houver"
  },
  "questoes_controvertidas": [
    "Fatos e pontos controvertidos"
  ],
  "provas": [],
  "rol_testemunhas": [],
  "advogados": [
    {
      "nome": "Nome",
      "OAB": "Número da OAB"
    }
  ]
}
### c) Réplica
{
  "tipo": { "key": 3, "description": "Réplica" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "peticionante": [
    { "nome": "Parte que apresenta a réplica" }
  ],
  "fatos": "Fatos novos ou reafirmações",
  "questoes_controvertidas": [
    "Pontos ainda controvertidos"
  ],
  "pedidos": [],
  "provas": [],
  "rol_testemunhas": [],
  "advogados": [
    {
      "nome": "Nome do advogado",
      "OAB": "OAB/UF Número"
    }
  ]
}
### d) Petição diversa
{
  "tipo": { "key": 6, "description": "Petição diversa" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "peticionante": [
    { "nome": "Parte que peticiona" }
  ],
  "causa_de_pedir": "Fatos e fundamentos da petição",
  "pedidos": [],
  "advogados": [
    {
      "nome": "Nome",
      "OAB": "OAB/UF Número"
    }
  ]
}
### e) Despacho inicial
{
  "tipo": { "key": 4, "description": "Despacho inicial" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "conteudo": {
    "desc": "Resumo do despacho"
  },
  "deliberado": [
    {
      "finalidade": "O que foi determinado",
      "destinatario": "Parte/autor/réu etc.",
      "prazo": "Prazo, se fixado"
    }
  ],
  "juiz": {
    "nome": "Nome do juiz"
  }
}
### f) Despacho ordinatório
{
  "tipo": { "key": 5, "description": "Despacho ordinatório" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "conteudo": {
    "desc": "Teor do despacho"
  },
  "deliberado": [
    {
      "finalidade": "Ex: intimação",
      "destinatario": "Parte ou advogado",
      "prazo": "Prazo fixado"
    }
  ],
  "juiz": {
    "nome": "Nome do juiz"
  }
}
### g) Decisão interlocutória / Tutela provisória
{
  "tipo": { "key": 7, "description": "Decisão interlocutória" },
  "processo": "Número do processo",
  "id_pje": "Conforme regra",
  "natureza": "Decisão interlocutória ou tutela provisória",
  "conteudo": [
    "Resumo da decisão e fundamentos"
  ],
  "deliberado": [
    {
      "finalidade": "Determinação ou concessão",
      "destinatario": "Parte/advogado",
      "prazo": "Prazo fixado"
    }
  ],
  "juiz": {
    "nome": "Nome do juiz"
  },
  "prazo": {
    "fixado": true,
    "detalhes": "Descrição do prazo"
  }
}


