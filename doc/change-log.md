# -----------------------------------------------------------------------------
#             Em 16/17-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) implementada a nova API da OpenAI Responses;
b) implementada a API REsponses para realizar chamadas de funções, a partir de
um prompt inserido pelo usuário;

# -----------------------------------------------------------------------------
#             Em 18-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) concluída a implementação da rotina genérica para utilização das novas API
Responses da openai, com a chamada de funções;
b) criadas rotinas genéricas para a criação e passagem das funções/parâmetros
das chamadas à API;
c) implementado o handler para as chamadas de análise dos autos;
d) iniciada a implementação da chamada da API na interface de análise do cliente;
e) reestruturada a aplicação backend para acomodar as rotinas de  tools e rag 
para uso das funcionalidade da openai.

# -----------------------------------------------------------------------------
#             Em 19-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) este changelog está partindo de um ponto já bastante avançado no desenvolvi-
mento do backend, mas se tornou necessários dar um tratamento em separado do de
senvolvimento do cliente, uma vez que as preocupações se tornaram bastante dis-
tintas e precisam de um registro mais preciso;

b) continuando as modificações no backend, foram implementadas melhorias nas 
rotinas para processamento RAG das requisições;
c) alterado o modelo para GPT-4.1, que apesar de se mostrar um pouco mais len-
mostrou uma maior qualidade na extração das informações das peças processuais,
o que irá melhorar a qualidade da minutas a serem geradas;
d) ajustados o teto máximo de token para as respostas do modelo, pois estava
muito baixo(512);
e) concentrei as funções a serem utilizadas pelo modelo no pacote rag, inclu-
sive a configuração do toolManager. 
f) passou-se a acrescentar ao prompt inserido pelo usuário a expressão 
"O contexto é igual a **", sendo a indicação do número do contexto trans-
parente para o usuário;
g) modifiquei o pacote "config" para excluir a variável de ambiente MaxTokens,
abandonada em favor da MaxCompletionTokens;
h) criado o endpoint "/contexto/autos/rag" para receber as solicitações
do cliente relacionadas à análise do contexto;

# -----------------------------------------------------------------------------
#             Em 20-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criada a variável de ambiente configuração "OPENAI_OPTION_MODEL_SECUNDARY" no
arquivo .env para manter um modelo secundário e mais econômico para ser usado
em atividades mais simples; também foi alterado o nome da variável de ambiente
do modelo principal, agora "OPENAI_OPTION_MODEL"
