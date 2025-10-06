### Fluxo do Agente de Análise Jurídica

|---------------------------|
|       Solicitação         |
|    Linguagem natural      |
|---------------------------|
            |   "contexto/query/rag"
            |       contextoQueryHandlers.QueryHandlerTools           
            |
            |-------------->    |---------------------------|
                                |  Chama o Orquestrador     |
                                |---------------------------|
                                            |
                                            |   StartPipeline(
                                            |        c.Request.Context(), 
                                            |        body.IdCtxt, 
                                            |        messages, 
                                            |        body.PrevID)
                                            |
                                |-----------------------|
                                |  Identifica o Evento  |
                                |-----------------------|
                                            |
                                            |   service.getNaturezaEventoSubmit(
                                            |       ctx, 
                                            |       idCtxt, 
                                            |       msgs, 
                                            |       prevID)
                                            |   service.handleEvento(
                                                    ctx, 
                                                    objTipo, 
                                                    id_ctxt, 
                                                    msgs, 
                                                    prevID)
                                            |
                                |-----------------------|
                                | Seleciona a Pipeline  |
                                |-----------------------|
                                            |
                                            |   service.pipelineAnaliseProcesso(
                                                    ctx, 
                                                    id_ctxt, 
                                                    msgs, 
                                                    prevID)
                                            |   service.pipelineProcessaSentenca(
                                                    ctx, 
                                                    id_ctxt, 
                                                    msgs, 
                                                    prevID)
                                            |   service.pipelineDialogoOutros(
                                                    ctx, 
                                                    id_ctxt, 
                                                    msgs, 
                                                    prevID)
                                            |   service.pipelineAddBase(
                                                    ctx, 
                                                    id_ctxt)
