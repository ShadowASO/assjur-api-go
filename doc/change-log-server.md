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

# -----------------------------------------------------------------------------
#             Em 26-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) incorporada as rotinas de extração de texto do arquivo PDF baixado do PJe
utilizando o aplicativo "pdftotext", por meio de chamada ao sistema;
b) alterado o Dockerfile para fazer a instalação do pacote "poppler-utils"
e modificada a imagem para "golang:1.24.4";
c) incorporadas as rotinas de divisão do arquivo texto gerado a partir da
extração do arquivo PDF gerado pelo PJe, salvando o conteúdo de cada docu-
mento em um registro do "docsocr";
d) criada rotina para fazer a análise de cada registro incorporado na tabe-
"docsocr" para verificar se é um documento aceitável para compor o acervo
processual, deletando os que não atenderem;
e) ajustada a interface do cliente para inserir um botão na janela Forma-
ção do Contexto para fazer a exclusão dos documentos inadequados;
f) ajustada a mesma interface para nao gerar um scrollbarr na janela prin-
cipal;

# -----------------------------------------------------------------------------
#             Em 30-06-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) alterada a denominação da tabela uploadfiles para uploads;
b) modificada a imagem do container golang de alpine para bulleyse, pois havia
incompatibilidade com o apt-get e tesseract;
c) 

# -----------------------------------------------------------------------------
#             Em 01-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criados aplicativos para fazerem o deploy do servidor e do cliente web no
servidor home, utilizando go lang. Para o server ficou o deploy-server e para
o cliente web ficou o deploy-web;

# -----------------------------------------------------------------------------
#             Em 02-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) realizado o upload do assjur na VPS;
b) problemas na configuração das variáveis de ambiente do .env; gerava erro na
conexão do postgres;
c) reconfigurados os arquivos do NGINX para aprimoramento da segurança;

# -----------------------------------------------------------------------------
#             Em 03-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) realizada limpeza nas imagens e volumes do Docker na VPS, pois estavam ocu-
pando muito espaço;
b) ajustes na interface de detalhes de modelos, testes no cadastro de modelos;

# -----------------------------------------------------------------------------
#             Em 04-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criados novos indices para guardarmos os embeddings de um processo e o inteiro
teor das decisões proferidas, permitindo a rápida busca semântica:
autos_embedding
decisões
b) criados os objetos de manipulação dos índices e do serviço chamado 
autosEmbedding. Ele irá gerar o embedding e utilizar os índices;

# -----------------------------------------------------------------------------
#             Em 05-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) continuando ajustes na API para trabalhar o embedding do processo como um todo;

# -----------------------------------------------------------------------------
#             Em 06-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) concluída as rotinas de inclusão de documento nos indices decisões e autos_embedding;
b) desenvolvendo as rotinas de formatação do json para gerar um embedding mais útil e 
significativo; já foi feito: criada rotina ParseJsonToEmbedding para identificar a natureza
do documento e chamar a rotina respectiva de parse; criadas as constantes de natureza dos
documentos: naturezaDocs.go; criado o parser para a petição inicial;

# -----------------------------------------------------------------------------
#             Em 07-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criados os indices "autos" e "autos_temp" no opensearch;
b) iniciada a migração dos documetnos do postgres para o opensearch;

# -----------------------------------------------------------------------------
#             Em 08-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) continuidade na conversão da API para utilizar indices do openSearch para o
salvamento dos documentos;
b) feitas alterações no cliente;

# -----------------------------------------------------------------------------
#             Em 09-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) desenvolvidas rotinas para extrair o índice do arquivo baixado do PJe para
selecionar os documentos que devem ser salvos no contexto, evitando o gasto de
tokens de IA para identificar os arquivos. Tornou mais rápido e econômico  a
importação das peças;
b) avançamos ainda mais na conversão das rotinas para o uso do banco vetorial
openSearch;
# -----------------------------------------------------------------------------
#             Em 10-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) concluída a migração de tabelas do postgresql para o banco vetorial openSearch;
b) concluída a revisão dos módulos do sistema e sua padronização;
c) modificada a URL;

# -----------------------------------------------------------------------------
#             Em 11-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) Criados 2 novos indices: autos_json_embedding e autos_doc_embedding;
b) Criado 2 novos módulos para manipular o indice autos_json_embedding:
autosJsonEmbedding e autosJsonService;

# -----------------------------------------------------------------------------
#             Em 14-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) concluída a análise do piperline de ingestão dos documentos do processo,
com a criação de parsers para a maior parte dos documentos.
b) criar um JSON para sentença e demais documentos que não tinham ainda;

# -----------------------------------------------------------------------------
#             Em 22-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) ajustadas as rotinas do serviço "rag" para utilizarem o indice autos no 
opensearch;
b) criada a natureza "NATU_DOC_ANALISE_IA" para identificar o documento relativo
à análise feita pela IA;
c) feitos diversos ajustes no cliente da aplicação para exibir o documento relativo
à última análise feita pela IA;
d) criado um módulo para o service do contexto;
e) feitas correções na API e na interface cliente;

# -----------------------------------------------------------------------------
#             Em 24-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) criação das rotinas para registrar o consumo de token por contexto;
b) ajustes nas rotinas de extração dos documentos do PJe;
c) alteração da interface cliente para exibir o número do processo, inseri botão
para excluir peça dos autos; 
d) vários testes;

# -----------------------------------------------------------------------------
#             Em 25-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) ajustes na interface do cliente para permitir selecionar os documentos a serem
deletados ou juntados por meio de checkbox; retirada dos botões de deleção e jun-
tada por documento;
b) alterada a interface de análise do processo para permitir selecionar cada peça
processual e permitir a deleção;
c) diversos testes de autuação e deleção de documentos do indice autos e do seu
embedding "autos_json_embedding". Modifiquei o campo doc_json para doc_json_raw,
guardando o objeto json como um string. O salvamento como um objeto estava geran-
do muitos erros, pois o primeiro campo determinava o tipo da propriedade e o sal-
vamenteo seguinte gerava erro;
e) muitos ajustes na interface e no servidor.

# -----------------------------------------------------------------------------
#             Em 26-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) ajustes finais nas rotinas de upload dos autos extraídos do PJe;
b) montagem do servidor Ryzen 7 5700G;

# -----------------------------------------------------------------------------
#             Em 27-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) configuração do servidor Ryzen 7 5700G;
b) feito o deploy do bando de dados postgres e do opensearch no servermaster;
c) configurado o Docker no novo servidor;

# -----------------------------------------------------------------------------
#             Em 28-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) feito o deploy da aplicação no servidor "servermaster" local;
b) feito o deploy da aplicação no servidor virtual(VPS), com recuperação do
backup do opernsearch e do postgresql;
c) alterado o aplicativo de deploy do servidor para fazer automatizar o procedi-
mento apenas com a informação de que o deploy é local ou na vps. O endereço do
host e o nome do arquivo de configuração será identificado automaticamente;
d) renomeadas as pastas da API da WEB para retirar referências à linguagem ou
framework;

# -----------------------------------------------------------------------------
#             Em 01-08-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) feitos diversos ajustes nas rotinas da API que operacionalizavam as ações 
RAG do sistema e realizavam análise e elaboração de sentenças. Feitas correções
que reduziram em 4x o gasto de tokens;
b) Melhorado o uso das tools functions com o modelo, que foi conciliado com o
uso das rotinas de prompt;
c) feitas várias correções na interface do cliente que tratavam do cadastro e 
modificação dos prompts;
d) pronto para os testes.

# -----------------------------------------------------------------------------
#             Em 02-08-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) alterado o módulo RAG para devolver respostas estruturadas, permitindo identi-
ficar a natureza da resposta, se uma análise, minuta de sentença ou simples bate
papo;
b) feitos ajustes na interface do cliente para exibir seletivamente as respostas
originadas na API RAG;
c) implementada na API e na interface do cliente a busca de contexto pelo número
do processo inserido parcialmente;

# -----------------------------------------------------------------------------
#             Em 03-08-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) excluído o módulo autosModel que atuava sobre a tabela autos do Postgresql;
b) tabela "autos" do Postgresql foi excluída;
c) tabela "docsocr" do Postgresql foi excluída;
d) alterada a documentação do schema de dados do Postgresql;
e) modificados os utilitários de deploy "deploy-web" e "deploy-api" para fazerem
a limpeza da respectiva pasta e deletar arquivos e pastas desnecessáriaas, tais 
como .git, node-modules, dist, server;
f) os utilitários foram melhorados para permitirem a derrubada dos containeres
e posterior levantamento dos containeres;

# -----------------------------------------------------------------------------
#             Em 04-08-2025: Versão 1.0.1                                  
# -----------------------------------------------------------------------------
a) corrigi a rotina de deleção do contexto e inseri uma verificação dos autos,
para não permitir a exclusão de um contexto de possua registros nos "autos";
b) inserido uma opção "Sobre" na janela principal do cliente, para exibir o
autor e a versão;
c) fiz ajustes na rotina ConsultaSemantica para permitir a busca em separado da
ementa_embedding e do inteiro_teor_embedding, com a natureza;

# -----------------------------------------------------------------------------
#             Em 09-08-2025: Versão 1.0.2                                 
# -----------------------------------------------------------------------------
a) criada API para obter a versão do servidor de API;
b) alterada a janela "Sobre" para exibir a versão do cliente e da API;
c) alterado o modelo de IA para o GPT-5-mini e feitos ajustes nas rotinas da 
API para ajustar a verbosidade e reasoning;
d) feitas alterações na interface para trabalhar com a API do novo modelo,
que trás o conteúdo significativo no registro [1] de output;

# -----------------------------------------------------------------------------
#             Em 10-08-2025: Versão 1.0.2                                 
# -----------------------------------------------------------------------------
a) inserido o consumo de tokens da interface de analise de processos;
b) feitas modificações no componente BarraListagem para que os botões fiquem
melhor visíveis; 
c) alterada a interface de ListaModelos e DetalheModelos para exibir o texto
formatado, com a inclusão de um novo compoenente utilizando o tiptap; modifi-
cados também para manter o estado de listagem quando há o retorno de detalhes
para listamodelos;
d) feita uma verdadeira reformulação em apiTools para tratamento de erros e 
utilização de funções helpers; 
e) feitos ajustes em ApiCliente para reforçar a segurança e consistência;
