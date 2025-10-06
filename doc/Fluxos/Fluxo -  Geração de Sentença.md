### Fluxo Geração de Sentença

|---------------------------------------|
|       Início do Pipeline              |
|       Processa Sentença               |
|---------------------------------------|
                    |
                    |   pipelineProcessaSentenca(
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
|-----------------------------------|
|  Recupera base de conhecimento    |
|-----------------------------------|
                |
                |   retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
                |
|---------------------------------------|
|  Executa a Analise de Julgamento      |
|---------------------------------------|
                |
                |   genObj.ExecutaAnaliseJulgamento(
                |       ctx, 
                |       id_ctxt, 
                |       msgs, 
                |       prevID, 
                |       autos, 
                |       ragDoutrina)
                |
                |---------------------> |---------------------------|
                                        | Seleciona um Prompt       |
                                        |   PROMPT_RAG_JULGAMENTO   |
                                        |---------------------------|      
                                                    |
                                                    |   services.PromptServiceGlobal
                                                    |   .GetPromptByNatureza(
                                                    |       consts.PROMPT_RAG_JULGAMENTO)
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
|  Salva a Minuta de Sentença       | 
|-----------------------------------|
                |
                |   service.salvarMinutaSentenca(
                |       ctx, 
                |       id_ctxt, 
                |       RAG_RESPONSE_SENTENCA, 
                |       resposta, "")
