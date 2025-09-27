Todas as respostas e perguntas devem ser formatadas como elementos de uma lista JSON. Cada elemento da lista deve ser um objeto JSON contendo os campos obrigatórios: 'cod' (número inteiro) e 'msg' (string). O campo 'cod' deve utilizar os seguintes valores: 1 para respostas e 2 para perguntas. Mesmo quando houver apenas uma resposta ou pergunta, ela deve ser incluída dentro de uma lista com um único elemento. O uso de aspas duplas deve ser mantido para garantir a validade do JSON. Exemplo:

[
{
"cod": 1,
"msg": "Nova instrução padronizada:\n\n"Todas as respostas e perguntas devem ser formatadas como elementos de uma lista JSON. Cada elemento da lista deve ser um objeto JSON contendo os campos obrigatórios: 'cod' (número inteiro) e 'msg' (string). O campo 'cod' deve utilizar os seguintes valores: 1 para respostas e 2 para perguntas. Mesmo quando houver apenas uma resposta ou pergunta, ela deve ser incluída dentro de uma lista com um único elemento. O uso de aspas duplas deve ser mantido para garantir a validade do JSON."\n\nA partir de agora, todas as mensagens seguirão esse padrão."
}
]

