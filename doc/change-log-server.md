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

# -----------------------------------------------------------------------------
#             Em 21-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) revisadas, corrigidas, padronizadas e melhoradas todos handlers da aplicação,
com a ajuda da IA;
b) criada a rotina "TokensCounter" para calcular a quantidade de tokens existen-
tes em um vetor de mensagens;
c) alterado o endpoint do chat para "/query/chat";
d) alterado o serviço "SubmitPromptResponse" para que ele possa receber uma 
string com o modelo a ser usado. Isso permitirá utilizar um modelo mais econô-
mico(mini) para ações rotineiras e deixar os modelo mais eficiente e caro para
as atividades que o exijam;
e) realizados testes na interfaces de análise do processo e a API tem se com-
portado muito bem, chamando todas as funções de peças do processo e fazendo
uma análise razoável do processo;

# -----------------------------------------------------------------------------
#             Em 22-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criei um aplicativo para separar os documentos constantes do arquivo completo
baixado do PJe(break-autos.go);
b) fiz o deploy da aplicação na atual situação para o servidor Home-srv;
c) modifiquei as rotinas de extração por OCR para permitir o upload de arquivos
.txt, sem precisar fazer OCR. Vai facilitar o manuseio de processo e testes;

# -----------------------------------------------------------------------------
#             Em 23-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criado aplicativo para extrair o texto de um arquivo PDF, usando OCR. Esse
arquivo deverá ser depois submetido ao aplicativo "autuar" que irá criar tan-
tos arquivos quanto sejam os documentos identificáveis no texto extraído;

b) Ajustada a rotina que faz o OCR no servidor para ignorar os arquivos com
extenão ".txt" e submetê-los à análise pela IA GPT-4.1-nano para verificar
se se enquadar em um dos seguintes documentos: petição inicial, contestação, 
réplica, despacho inicial, despacho ordinatório, petição diversa, 
decisão	interlocutória, sentença, embargos de declaração, contra-razões, 
apelação ou laudo pericial;

c) realizados alguns testes, sendo constatados alguns erros da IA, tais como
tratando certidões ou reproduções das peças como as próprias;

# -----------------------------------------------------------------------------
#             Em 24-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) realizados ajustes nos endpoints da API para que o tratamento por OCR seja
apontado por "/contexto/documentos/ocr";
b) desmembrei a rotina de extração para permitir a execução pela indicação dos
documentos a serem extraídos e crie um endpoint "/contexto/documentos/ocr/:id"
para acionar a extração para todos os documentos de um contexto;
c) inseri um botão na janela de Formação do Contexto Processual para permitir
a extração de todos os documentos trasnferidos por upload e vinculados a um 
determinado contexto;

# -----------------------------------------------------------------------------
#             Em 25-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) desenvolvi duas novas ferramentas para fazer o tratamento do arquivo PDF 
extraído por download do PJe;
b) o utilitário "pdftotext" faz a extração do texto contido no PDF e cria
um novo aquivo com o mesmo nome e a extensão .txt;
c) o utilitário "pdfautos" trabalha no arquivo gerado pelo "pdftotext", 
criando uma pasta chamada "Autos" e um arquivo para cada documento dos autos;
Obs. Os resultados usando os dois novos utilitário ficaram muito melhor do 
que usando OCR.


