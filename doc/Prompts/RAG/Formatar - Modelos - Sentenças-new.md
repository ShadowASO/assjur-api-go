Prompt Enxuto – Extração Estruturada de Sentenças

Extraia e estruture o conteúdo da sentença judicial no formato JSON válido:

id_pje: número de 9 dígitos do rodapé “Num. ######### - Pág.”; se houver mais de um, use o último; se não houver, "id_pje não identificado".

Fidelidade máxima: copiar texto integral, sem resumir ou reescrever.

Questões: cada preliminar e cada mérito em objeto separado da lista "questoes".

Parágrafos: dividir integralmente em vetor de strings.

Decisão: síntese curta (ex.: "Rejeitada", "Procedente, fixado em R$ 5.000,00").

Dispositivo: transcrever integralmente em vetor de strings.

Ausência de informação: usar "".

Saída final: apenas JSON válido.

{
  "tipo": {
    "key": 7,
    "description": "Sentença judicial"
  },
  "processo": "",
  "id_pje": "",
  "metadados": {
    "classe": "",
    "assunto": "",
    "juizo": "",
    "partes": {
      "autor": "",
      "reu": ""
    }
  },
  "questoes": [
    {
      "tipo": "preliminar",
      "tema": "",
      "paragrafos": [],
      "decisao": ""
    },
    {
      "tipo": "mérito",
      "tema": "",
      "paragrafos": [],
      "decisao": ""
    }
  ],
  "dispositivo": {
    "paragrafos": []
  }
}

