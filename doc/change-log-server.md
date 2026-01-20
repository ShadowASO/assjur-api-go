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
b) feito o deploy do banco de dados postgres e do opensearch no servermaster;
c) configurado o Docker no novo servidor;

# -----------------------------------------------------------------------------
#             Em 28-07-2025: Versão 1.0.0                                    
# -----------------------------------------------------------------------------
a) feito o deploy da aplicação no servidor "servermaster" local;
b) feito o deploy da aplicação no servidor virtual(VPS), com recuperação do
backup do opernsearch e do postgresql;
c) alterado o aplicativo de deploy do servidor para  automatizar o procedi-
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
para não permitir a exclusão de um contexto que possua registros nos "autos";

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

f) Ajustes geral nas interfaces para mostrar uma barra superior com a descri-
ção da janela e um botão para abrir e fechar o menu lateral; muita reestilização;

g) feito o deploy da versão 1.0.2 no servidor-home e na VPS;

# -----------------------------------------------------------------------------
#             Em 11-08-2025: Versão 1.1.0                                 
# -----------------------------------------------------------------------------
a) reformulado o módulo de autenticação, com atualização do pacote jwt para
"github.com/golang-jwt/jwt/v5";

b) restrição de acesso aos módulos de processos e prompts;

c) segregação das rotas do main para um package próprio chamado "rotas";

d) feitos ajustes nos arquivos .env de configuração de ambiente para retirar
aspas;

e) ajustado o cliente web para que o título da janela Análise de Processo 
fosse atribuída nos efeitos;

f) revisado o código do config.go com o GPT-5;

#************************************
g) modificados os utilitários de deploy e criado um novo utilitário chamado
"deploy-assjur" que permite fazer do deploy individual das cadasdas ou ambos
ao mesmo tempo; foi também solucionado o problema de solicitação repetitiva
de senhas; ficou muito mais rápido o deploy; realizados testes no servidor
local e na VPS;
#**********************************

h) aumentado o número de registros devolvidos nas buscas semânticas para 10
registros;

# -----------------------------------------------------------------------------
#             Em 12-08-2025: Versão 1.1.1                                 
# -----------------------------------------------------------------------------
a) desabilita o combobox da natureza quando não for o cadastro de um novo modelo;

b) reestilizei a BarraDetalhes para reagir de acordo com o modo em que
estiver: view, edit, create;

c) aplicadas as alterações em "DetalheModelos" e "DetalhePrompt";

d) feita a exclusão de pastas que continham código remanescente do curso de
MUI;

e) refatorado o logger com o GPT-5-mini;

f) feito o DEPLOY do cliente WEB na VPS;

# -----------------------------------------------------------------------------
#             Em 13-08-2025: Versão 1.1.2                               
# -----------------------------------------------------------------------------
a) componentes FlashProvider e FlashAlerta foram refatorados pelo GPT-5-mini, que
corrigiu erros e fez melhorias significativas;

b) sistema exibe flash mensagem informando que o usuário não possui permissão ou 
qualquer outro erro de forma amigável;

c) feita correção na API para tratar o Output do modelo no pipeline;

d) feitas diversas alterações nas rotinas que fazem a extração dos documentos
dos autos, a partir do arquivo txt gerado pela conversão do PDF. Alguns docu-
mentos não estavam sendo identificados;

e) correção dos componentes ListaDocumentos e UploadProcesso para correção e erros
e melhoria. Foi corrigido um erro que duplicava o comando de autuação dos docu-
mentos;

f) aperfeiçoada a janela modal de criação de contextos;

g) modificado o SystemProvider para expor os estados versionApi e versionApp;

# -----------------------------------------------------------------------------
#             Em 14-08-2025: Versão 1.1.3                              
# -----------------------------------------------------------------------------
a) Aperfeiçoada a interface da janela de Bate-Papo e a janela de Análise do Pro-
cesso, colocando um spinner;

b) Alterada a janela de análise de processo para fazer um refressh dos autos
quando houve uma análise;

c) alterado o servidor de API para o WriteTimeout para WriteTimeout: 5 * time.Minute, pois o tempo an-
terior de 60 segundos estava muito curto;

d) criado um ClientGoneMiddleware para tratar

e) modifiquei o prompt de análise de pelas processuais para uma versão mais 
enxuta criada pelo GPT-5. Isso reduziu a quantidade de tokens de entrada em 793
e na saída em 11.884;

# -----------------------------------------------------------------------------
#             Em 15-08-2025: Versão 1.1.4                             
# -----------------------------------------------------------------------------
a) ajustes na API da openapi.go e openaiService.go

# -----------------------------------------------------------------------------
#             Em 17-08-2025: Versão 1.1.4                             
# -----------------------------------------------------------------------------
a) ajustes na estrutura do código;

# -----------------------------------------------------------------------------
#             Em 23-08-2025: Versão 1.1.4                             
# -----------------------------------------------------------------------------
a) feito o DEPLOY da versão 1.1.4 no servidor home e na VPS;

# -----------------------------------------------------------------------------
#             Em 23-08-2025: Versão 1.1.5                           
# -----------------------------------------------------------------------------
a) ajustes no nome do source código de main.go para server.go;
b) Compilação: go build -v -o server ./cmd/server.go
   Execução:   ./server
c) modificado o tamanho máximo do documento processual de 60kb para 180kb;
d) modificada a rotina isDocumentoSizeValido para ignorar documentos sem conteúdo,
caracterizado por conter apenas um alinha de texto;
e) alterada a ordem dos botões na janela dialog que exibe o conteúdo do documento,
para fixar o botão de exclusão à esquerda, para evitar ações acidentais.

# -----------------------------------------------------------------------------
#             Em 10/19-09-2025: Versão 1.1.6                          
# -----------------------------------------------------------------------------
a) conclusão da aplicação com RAG
b) conclusão do TCC

# -----------------------------------------------------------------------------
#             Em 27-09-2025: Versão 1.1.7                        
# -----------------------------------------------------------------------------
a) paginação do select de contextos;
b) atualização do servidor VPS e home;

# -----------------------------------------------------------------------------
#             Em 27-09-2025: Versão 2.0.0                        
# -----------------------------------------------------------------------------
# TODO
# Aperfeiçoar das rotinas RAG

# -----------------------------------------------------------------------------
#             Em 29-09-2025: Versão 2.0.0                        
# -----------------------------------------------------------------------------
a) ajustes no prompt de análise jurídica e mudança na estratégia de retorno
dos dados no formato JSON para manipulação pelo cliente;
b) ajustes no prompt da geração de minutas de sentença, com adaptação do código;
c) criado um objeto JSON para sentenças;

# -----------------------------------------------------------------------------
#             Em 30-09-2025: Versão 2.0.0                        
# -----------------------------------------------------------------------------
a) ajustes no prompt de geração de sentenças para que os parágrafos do relatório
e da fundamentação de mérito venham separados em strings de um vetor para pos-
bilitar manipular no código;
b) alterada a denominação das constantes RAG_SUBMIT p/RAG_EVENTO;

# -----------------------------------------------------------------------------
#             Em 01-10-2025: Versão 2.0.0                        
# -----------------------------------------------------------------------------
a) criado o índice "RAG_DOC_EMBEDDING" para guardar os fragmentos das sentenças;
b) criados o handler e o serviço para manipular o "RAG_DOC_EMBEDDING" no servidor;
c) criado um novo item no menu lateral chamado RAG e uma nova interface para lis-
tar e modificar a base de conhecimentos contida o índex "RAG_DOC_EMBEDDING".

d) criadas as constantes de classe e assunto com base nas tabelas no CNJ;
e) teste de inclusão e busca semântica corretas;
f) criado o prompt de extração da sentença proferida nos autos do processo.

# -----------------------------------------------------------------------------
#             Em 02-10-2025: Versão 2.0.0                        
# -----------------------------------------------------------------------------
a) concluído ajuste no pipeline de extração de documentos para tratar a senten-
ça utilizando um prompt de extração de sentenças;
b) geração do objeto json com sucesso e exibindo na janela Analise Jurídica;
c) iniciado o pipeline para a ingestão da sentença e salvá-la no índice RAG;
d) ajustado o prompt de formatação da sentença para quebrar as várias questões
de mérito tratadas em uma sentença;

# -----------------------------------------------------------------------------
#             Em 05/06-10-2025: Versão 2.0.2                       
# -----------------------------------------------------------------------------
a) ajustada a nomenclatura dos pacotes que tratam o "rag_doc_embedding";
b) documentados os fluxos de trabalho;
c) corrigido o package que tratam o "rag_doc_embedding";
d) concluída rotinas de salvamento dos tópicos da sentença;

# -----------------------------------------------------------------------------
#             Em 07/06-10-2025: Versão 2.0.2                       
# -----------------------------------------------------------------------------
a) ajustes no prompt Análise Jurídica para formular perguntas ao usuários em 
relação às questões controvertidas;
b) sistema exige a realização de uma pré-análise do processo e faz a verifica-
ção das questões controvertidas quando o usuário solicita a geração de uma mi-
nuta de sentença;
c) criado um prompt chamado "Valida - Completude - SEntença" para verificar se
existem questões controvertidas a serem decididas pelo usuário;
d) modificada a denominação dos prompts;
e) ajustada a interface do cliente para exibir as questões controvertidas na
janela de diálogo;
f) vários testes;

# -----------------------------------------------------------------------------
#             Em 08-10-2025: Versão 2.2.0                       
# -----------------------------------------------------------------------------

# -----------------------------------------------------------------------------
#             Em 09-10-2025: Versão 2.0.2                       
# -----------------------------------------------------------------------------
a) feita a separação dos eventos gerados dos documentos dos autos, com a criação
do índice "eventos" e do eventosIndex, eventosServices e eventosHandler;
b) adaptada a interface para utilizar o novo índice;
c) muitas mudanças na interface do cliente, com criação de novos componentes;

# -----------------------------------------------------------------------------
#             Em 10-10-2025: Versão 2.0.2                       
# -----------------------------------------------------------------------------
a) criada a função REcuperaBaseConhecimento para fazer a recuperação da base
de conhecimento, utilizando recursos de concorrência;
b) vários testes na interface do cliente;

# -----------------------------------------------------------------------------
#             Em 11-10-2025: Versão 2.3.0                     
# -----------------------------------------------------------------------------
a) concluída a formatação da minuta de sentença, bem como a geração de PDF e
impressão;
b) ajustada a interface da janela Análise Jurídica para aproveitar melhor o espaço
vertical;
c) concluídas as rotinas de cópia, impressão e geração de PDF relativo às
minutas;

# -----------------------------------------------------------------------------
#             Em 12-10-2025: Versão 2.3.0                     
# -----------------------------------------------------------------------------
a) fiz uma ampla formatação da análise jurídica e da sentença;
b) fiz ajustes nas rotinas de impressão, cópia e criação de arquivo pdf;
c) atualizei o servidor local e a VPS;
d) configurei o nginx da VPS para aceitar upload de até 40m;

# -----------------------------------------------------------------------------
#             Em 13-10-2025: Versão 2.3.1                     
# -----------------------------------------------------------------------------
a) amplicação para os tipos de natureza de documentos que o sistema pode extrair: 
termo de audiência, ata de audiência, embargos de declaração, alegações finais, 
informações, proposta de acordo, ;
b) corrigi o parágrafo de conclusão da sentença;
c) ampliado para autuar embargos, parecer do MP,  laudo pericial e rol de tes-
temunhas;
d) retirei a autorização de admin para acesso aos contextos;
e) feito o deploy no servidor home.
f) feito o deploy na VPS.

# -----------------------------------------------------------------------------
#             Em 14-10-2025: Versão 2.3.2                    
# -----------------------------------------------------------------------------
a) retirar a verificação do processo no CNJ como condição para a criação de um
novo contexto/processo; estou adontando um modelo temporário com dados genéricos
nos casos em que a API do CNJ não retorna nada, como está acontecendo atualmente;
b) inseri verificação na inclusão de sentença na base de conhecimentos;
c) ajustes no prompt de identificação de eventos e inteções do usuário;
d) ajustes nos logs de mensagens do servidor;

# -----------------------------------------------------------------------------
#             Em 14-10-2025: Versão 2.3.3                   
# -----------------------------------------------------------------------------
a) validação do número do processo por meio de algorítimo;

# -----------------------------------------------------------------------------
#             Em 15-10-2025: Versão 2.4.0                   
# -----------------------------------------------------------------------------
a) ajustes no prompt de Extração das peças dos autos para modificar o formato
do json do Laudo Pericial;
b) extensas modificações nos prompts de extração de peças processuais para in-
cluir a data da assinatura eletrônica e o nome de quem a assinou. Essa data
equivale à data da inclusão nos autos;
c) inclusão de data de geração na análise jurídica;
d) feito o upload no home e vps;

# -----------------------------------------------------------------------------
#             Em 15-10-2025: Versão 2.4.1                   
# -----------------------------------------------------------------------------
a) inserido botão na janela Formação do Contexto para o usuário ir direto para
a janela Análise Jurídica;

# -----------------------------------------------------------------------------
#             Em 16-10-2025: Versão 2.4.1                   
# -----------------------------------------------------------------------------
a) feito upgrade na VPS e reconfigurada toda a parte de segurança;

# -----------------------------------------------------------------------------
#             Em 18-10-2025: Versão 2.4.2                  
# -----------------------------------------------------------------------------
a) atualizada a biblioteca oficial para acesso a OpenAPI para v3(3.5);
b) corrigido o retorno dos eventos RAG_EVENTO_ADD_SENTENCA, tanto na API quan-
to no cliente. Na API passou a retornar um []responses.ResponseOutputItemUnion;
c) criada rotina para devolver um responses.ResponseOutputItemUnion vazio;
d) alterado o json do prompt outros para retornar "conteudo" no lugar de texto;
e) ajustes na interface Analise Jurídica para exibir scrollbar vertical no
dialog;
f) ajustada a formatação do documento minuta de sentença e análise jurídica
para exibir a data de geração;
g) corrigido um erro no tipo da MinutaSentença no servidor;
h) DEPLOY v2.4.2 - Com modificações;
i) corrigido problema de timezone do docker compose;

# -----------------------------------------------------------------------------
#             Em 19-10-2025: Versão 2.4.3                  
# -----------------------------------------------------------------------------
a) modificada a função VerificaQuestõesControvertidas para utilizar as análises
jurídicas como fonte das questões controvertidas, e não a pré-análise jurídica;

# -----------------------------------------------------------------------------
#             Em 20-10-2025: Versão 2.4.4                  
# -----------------------------------------------------------------------------
a) corrigoda rotina de extração das peças processuais para reconhecer ata de 
audiência de conciliação;
b) corrigido erro na API que duplicava o retorno de mensagem de erro na rotina
RAG, quando não havia análise jurídicas;
c) ajustas as pipelines de análise jurídica e análise de julgamento, refinando
os prompts developer e alterando o role para o prompt de análise e de julgamento
para developer;
d) criadas rotinas auxiliares para a execução da análise e julgamento;

# -----------------------------------------------------------------------------
#             Em 21-10-2025: Versão 2.4.6                
# -----------------------------------------------------------------------------
a) saneamento nas rotinas utilizadas nas técnicas de RAG, com otimização do có-
go e uniformização das rotinas utilitárias;
b) feito o comentário do código e alterada a nomenclatura de algumas funções;
c) criado componente AltrarContexto no cliente web, permitindo a alteração do Juízo,
a Classe e os Assuntos;
d) UPLOAD versão 2.4.6 para os servidores home e VPS.

# -----------------------------------------------------------------------------
#             Em 22-10-2025: Versão 2.4.7                 
# -----------------------------------------------------------------------------
a) ocultada a aba RENDERIZADA da janela Análise Jurídica;
b) modificado o servidor para extrair os docuementos "Petição intermediária" e
"Rol de Testemunhas";
c) corrigida janela Análise Jurídica para atualizar o consumo de tokens no
retorno dos eventos;
d) ajuste a endpoint do consumo de tokens;

# -----------------------------------------------------------------------------
#             Em 23-10-2025: Versão 2.4.8                 
# -----------------------------------------------------------------------------
Alterações no cliente WEB
a) correção na janela Alterar Contexto, que estava com a classe e o juízo troca-
dos;
b) criado um compoenente específico para o prompt, pois estava gerando uma exces-
siva renderização com lentidão na digitação; criado o compoenten PromptInput;
c) modificada a coluna 01, dos autos e minutas, para que não ultrapassasse a al-
tura da janela, exibindo um scrollbar na janela principal. Aproveitei para pa-
dronizar as duas partes da coluna, retirando o Accordion das minutas;
d) retirado o ícono pdf para geração de um arquivo com a minuta; o salvamento
poderá ser feito no momento da impressão;

# -----------------------------------------------------------------------------
#             Em 24-10-2025: Versão 2.5.0                 
# -----------------------------------------------------------------------------
a) invertida a lógica de extração das peças processuais para definir os tipos 
que estão excluídos, e não os tipos autorizados. Isso evita a desconsideração
de peças e documentos relevantes, mas cadastrados incorretamente no PJe;

# -----------------------------------------------------------------------------
#             Em 24-10-2025: Versão 2.5.1                 
# -----------------------------------------------------------------------------
a) identifica os erros de violação da política da OpenAI. Inserido tratamento
na biblioteca e atualizado a biblioteca do cliente openai 3.6.1;
# -----------------------------------------------------------------------------
#             Em 24-10-2025: Versão 2.5.2                 
# -----------------------------------------------------------------------------
a) upload timeout elevado para 5 minutos;
b) correção nos prompts "extrair-autos" e "extrair - sentença" para extrair o 
id_pje nos casos que ele tem menos de 9 dígitos;
c) alterada API de listagem de peças para exibir até 50 registros;
d) alterada a janela Dialog que exibe as peças processuais extraídas do PDF.
Inseridos botões para avanço e retrocesso, e quando houve a deleção de um
registro.
e) ajustes na função SubmitPromptResponse_openai para evitar erros panics;

# -----------------------------------------------------------------------------
#             Em 25-10-2025: Versão 2.5.3                 
# -----------------------------------------------------------------------------
a) aplicado destaque nos registros selecionados na janela Análise Jurídica e na
lista de peças da janela Formação do Contexto;
b) modificação da janela Dialog para que se comporte como painel deslizante e 
localizada à esquerda, sem encobrir os registros de documentos extraídos;

# -----------------------------------------------------------------------------
#             Em 31-10-2025: Versão 2.5.4                 
# -----------------------------------------------------------------------------
a) aumentado o tamanho máximo de upload para 80MB e criada a constante MAX_SIZE_UPLOAD;

b) DEPLOY no serverhome;

# -----------------------------------------------------------------------------
#             Em 01-11-2025: Versão 2.5.4                 
# -----------------------------------------------------------------------------
a) Atualizado no GitHub;
b) DEPLOY no homeserver e VPS;

# -----------------------------------------------------------------------------
#             Em 04-11-2025: Versão 2.5.4               
# -----------------------------------------------------------------------------
a) alterado o prompt - extrair - autos para inserir o campo "fatos" no json da
petição inicial e "versão_dos_fatos" na json da contextação;

# -----------------------------------------------------------------------------
#             Em 01-12-2025: Versão 2.5.6               
# -----------------------------------------------------------------------------
a) alterada a estrutura da tabela uploads, no postgres, para inserir o campo
dt_inc com o tipo timestamp, permitindo guardar data e time da inclusão. Isso
possibilitará a exclusão automática, em um tempo a ser definido, dos arquivos
transferidos por upload e esquecidos.

b) altrado o índice autos_temp, no opensearch, para inserir o campo dt_inc.
Isso permitirá o expurgo automático dos registros extraídos, após um determi-
nado período de tempo;

c) feitas alterações no código de inclusão dos registros na tabela uploads e
autos_temp para tratar o novo campo.

# -----------------------------------------------------------------------------
#             Em 31-12-2025: Versão 3.0.0               
# -----------------------------------------------------------------------------
a) concluída a reestruturação das bases do sistema, com a transferência da tabela
de contextos para o OpenSearch, concentrando o contexto num único banco de dados,
o que permitirá fazer backups e restaurá-lo sem problemas de consistência;
b) alterada a sistemática de criação do valor do campo id_ctxt que deixou de
ser um inteiro gerado automaticamente pelo banco de dados e passou a ser gera-
do no próprio servidor, utilizando um string uuid-v7;

# -----------------------------------------------------------------------------
#             Em 01-01-2026: Versão 3.0.0               
# -----------------------------------------------------------------------------
a) ajustes no módulo "contextoIndex" para padronizar o uso da API do opensearch
client;

# -----------------------------------------------------------------------------
#             Em 02-01-2026: Versão 3.0.0               
# -----------------------------------------------------------------------------
a) feitos diversos ajustes nas rotinas de manipulação do opensearch, principal-
mente no sentido de padronizar a codificação;

# -----------------------------------------------------------------------------
#             Em 07-01-2026: Versão 3.1.1              
# -----------------------------------------------------------------------------
a) correção de vários erros nas rotinas de opensearch, com padronização de có-
digo e melhorias na interface;
b) aperfeiçoamento no prompt de identificação de intenções, com uma melhoria
significativa na interação com o usuário;
c) melhoria no prompt de análise jurídica;
d) feito o deploy da versão na VPS e no servidor home;

# -----------------------------------------------------------------------------
#             Em 08-01-2026: Versão 3.1.2             
# -----------------------------------------------------------------------------
a) ajustes no prompt Prompt - Análise Jurídica, agora na versão 
(Prompt - Análise Jurídica(V3);
b) ajustes na interface de modelos;
c) ajustes nos componentes BarraListagem e BarraDetalhes;

# -----------------------------------------------------------------------------
#             Em 09-01-2026: Versão 3.2.0             
# -----------------------------------------------------------------------------
a) alterada a estrutura de "base_doc_embedding" para acrescentar os campos 
id_ctxt, username_inc, dt_inc, hash_texto. Alteramos os nomes dos campos
para texto e texto_embedding;
b) muitas alterações nas rotinas de handler, service e index do índice
"base_doc_embedding", inclusive corrigimos um uso excrescente do índice,
diretamente pelas rotinas de ingestão;
c) alteradas as rotinas de analise jurídica, geração de minutas para inserir
o nome do usuário e o id_ctxt;
d) correções nas rotinas do clientes;
e) alterada a estrutura do index "eventos" para incluir os campos 
username_inc e dt_inc;

# -----------------------------------------------------------------------------
#             Em 10-01-2026: Versão 3.2.1             
# -----------------------------------------------------------------------------
a) ajustes finais das rotinas de formação da base de conhecimento, com adoção
de hash do texto e verificação de existência na adição da sentença à base de
conhecimentos;

# -----------------------------------------------------------------------------
#             Em 10-01-2026: Versão 3.2.2             
# -----------------------------------------------------------------------------
a) alterada a denominação das rotas "rag" para "base";
b) inserido parâmetro "promptCacheKey" nas chamadas à API da OpenAI, criando
um cache de 24 horas;
c) corrido erro nas inclusões de documentos na base de conhecimento, onde
era enviada a chave "id_ctxt" erroneamente;

# -----------------------------------------------------------------------------
#             Em 11-01-2026: Versão 3.2.4           
# -----------------------------------------------------------------------------
a) alterado o pipeline de análise para que as diversas rotinas devolvam um objeto
mais rico em informações sobre erros;
b) alteração na resposta da API "/contexto/query/analise" para devolver uma res-
posta minunciosa, com informações sobre erros ocorridos durante uma análise e 
eventuais pendências;
c) feitas diversas modificações no código do frontend para adequar à nova resposta
da API "analise";
d) ajustes no tratamento das respostas oriundas de busca pelo ID;

# -----------------------------------------------------------------------------
#             Em 12-01-2026: Versão 3.2.5           
# -----------------------------------------------------------------------------
a) ADOÇÃO DO MODELO MAIS AVANÇADOS GPT-5.2 PARA ANÁLISE E GERAÇÃO DE MINUTAS;
b) criada a variável de ambiente OPENAI_OPTION_MODEL_TOP='gpt-5.2' no arquivo
.env para que o modela mais avançado possa ser utilizado em situações especí-
ficas em que a análise aprimorada é necessária, tal como na análise jurídica e
na geração da minuta de sentença;
c) alterei as chamadas 'services.OpenaiServiceGlobal.SubmitPromptResponse()'
nas funções "ExecutaAnaliseJulgamento" e "ExecutaAnaliseProcesso" para usa-
rem o modelo TOP. Avaliar a disponibilização na VPS;

# -----------------------------------------------------------------------------
#             Em 14-01-2026: Versão 3.2.10          
# -----------------------------------------------------------------------------
a) alterada o servidor para que a busca na base de conhecimentos seja realizada
corretamente, pelo conteúdo e não pelo simples tema;

b) inserido o campo "base" na análise jurídica para exibir as informações na
base de conhecimento;

c) adotamos o modelo GPT5.2 para a geração das minutas de sentenças. A análise
permaneceu no GPT5-mini;

# -----------------------------------------------------------------------------
#             Em 17-01-2026: Versão 3.3.0          
# -----------------------------------------------------------------------------

a) alterado o menu vertical da página inicial para alterar a denominação de 
Página inicial para Início; Processos para Contextos; RAG para Precedentes; 
e excluí o item Bate-papo;

# -----------------------------------------------------------------------------
#             Em 18-01-2026: Versão 3.3.0          
# -----------------------------------------------------------------------------
a) Revisão do TCC;
b) Alteração na nomenclatura das janelas e nos itens de menu para conferir uni-
formidade e coerência;

# -----------------------------------------------------------------------------
#             Em 18-01-2026: Versão 3.3.1          
# -----------------------------------------------------------------------------
a) modificada a interface de formação do contexto para ocultar "Outros Documentos",
inserindo um checkbox por meio do qual o usuário pode mudar o comportamento e exi-
bir tudo.

# -----------------------------------------------------------------------------
#             Em 18-01-2026: Versão 3.3.2          
# -----------------------------------------------------------------------------
a) criado serviço de limpeza do índice "autos_temp", que roda a cada hora e de-
leta os registros inseridos há mais de 24 horas;
b) modificado o checkbox para Listar Todos;
