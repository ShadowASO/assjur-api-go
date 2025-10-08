### Fluxo da autuação de documentos

|----------------|
|    API         |
|----------------|
        |   "/contexto/documentos/autua"
        |       AutuaDocumentosHandler
                            |
                |---------------------------|
                |   ProcessarDocumentos     |
                |---------------------------|
                            |
                            |   services.ProcessarDocumento(idCtxt, idDoc)
                            |        
                    |---------------------------|
                    | Selecionar os documentos  |
                    |    em "autos_temp"        |
                    |---------------------------|
                                |
                                |   AutosTempServiceGlobal.SelectById(IdDoc)
                                |   
                    |---------------------------|
                    | Verificar duplicidades    |
                    |---------------------------|
                                |
                                |   AutosServiceGlobal.IsDocAutuado(IdContexto, row.IdPje)
                                |
                    |-------------------------------|       |-----------------------|
                    |       Selecionar um Prompt    |  ---> | Sentenças usa Prompt  |
                    |-------------------------------|       |       próprio         |
                                |                           |-----------------------|
                                |
                                |   PromptServiceGlobal.GetPromptByNatureza(natuPrompt)
                                |
                    |---------------------------|
                    |   Submeter ao Modelo      |
                    |       Gera JSON           |
                    |---------------------------|
                                |
                                |   OpenaiServiceGlobal.SubmitPromptResponse
                                |
                    |---------------------------|
                    |   Atualiza tokens no      |
                    |       contexto            |
                    |---------------------------|
                                |
                                |   ContextoServiceGlobal.UpdateTokenUso(
                                |       IdContexto, 
                                |       int(usage.InputTokens), 
                                |       int(usage.OutputTokens))
                                |
                    |-------------------------------|
                    |   Salva o json em "autos"     |
                    |-------------------------------|
                                |
                                |   rowAutos, err := AutosServiceGlobal.InserirAutos(
                                |        idCtxt, 
                                |        idNatu, 
                                |        idPje, 
                                |        row.Doc, 
                                |        rspJson)
                                |
                    |-----------------------------------|
                    |   Gera embedding do JSON          |
                    |   Salva em "autos_json_embedding" |
                    |-----------------------------------|
                                |
                                |   ialib.GetDocumentoEmbeddings(jsonRaw)
                                |   AutosJsonServiceGlobal.InserirEmbedding(
                                |       rowAutos.Id, 
                                |       idCtxt, 
                                |       idNatu, 
                                |       embVector)
                                |
                    |-----------------------------------|
                    |  Deleta doc em "autos_temp"       |
                    |-----------------------------------|
