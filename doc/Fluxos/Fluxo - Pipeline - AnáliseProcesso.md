### Fluxo do Pipeline AnáliseProcesso

|---------------------------------------|
|       Início do Pipeline              |
|       Análise do Processo             |
|---------------------------------------|
                    |
                    |   pipelineAnaliseProcesso(
	                |       ctx context.Context,
	                |       id_ctxt int,
	                |       msgs ialib.MsgGpt,
	                |       prevID string)
                    |  
|---------------------------------------|
|   Recupera os documentos dos autos    |
|---------------------------------------|
                    |
                    |    retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
                    |
|---------------------------------------|
|       Recupera Pré-análise            |
|---------------------------------------|
                |
                |   retriObj.RecuperaPreAnaliseJudicial(ctx, id_ctxt)
                |   
                | if existe pré-analise
|-----------------------------------|
|  Recupera base de conhecimento    |
|-----------------------------------|
                |
                |   retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
                |
|---------------------------------------|
|  Executa a Analise do Processo        |
|---------------------------------------|
                |
                |   genObj.ExecutaAnaliseProcesso(
                |       ctx, 
                |       id_ctxt, 
                |       msgs, 
                |       prevID, 
                |       autos, 
                |       ragDoutrina)
                |
                |---------------------> |-----------------------|
                                        | Seleciona um Prompt   |
                                        |   PROMPT_RAG_ANALISE  |
                                        |-----------------------|
                                                    |
                                                    |   services.PromptServiceGlobal
                                                    |   .GetPromptByNatureza(
                                                    |       consts.PROMPT_RAG_ANALISE)
                                                    |
                                        |---------------------------|
                                        |   Submete ao Modelo       |
                                        |---------------------------|
                                                    |
                                                    |   services.OpenaiServiceGlobal
                                                    |   .SubmitPromptResponse(
                                                    |       ctx,
                                                    |       messages, 
                                                    |       prevID,
                                                    |       config.GlobalConfig
                                                    |           .OpenOptionModel,
                                                    |       ialib.REASONING_LOW,
                                                    |       ialib.VERBOSITY_LOW)
                                                    |
                                        |---------------------------|
                                        |   Atualiza o Consumo      |
                                        |           de              |
                                        |         tokens            |  
                                        |---------------------------|
                                                    |
                |<--------------------------------- |
|-----------------------------------|
|  Salva o Resultado da Analise     | 
|-----------------------------------|
                |
                |   service.salvarAnaliseProcesso(ctx, id_ctxt, natuAnalise, resposta, "")
