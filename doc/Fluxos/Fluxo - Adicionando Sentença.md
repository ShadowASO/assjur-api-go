### Fluxo Adicionando Sentença à Base de Conhecimento

|---------------------------------------|
|       Início do Pipeline              |
|       Adição de Sentença              |
|              à                        |
|       Base de Conhecimento            |
|---------------------------------------|
                    |
                    |   pipelineAddBase(
	                |       ctx context.Context,
	                |       id_ctxt int)
                    |  
|---------------------------------------|
|   Recupera a Sentença dos autos       |
|---------------------------------------|
                    |
                    |    retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
                    |    
|---------------------------------------|
|   Inicia a inclusão da Sentença       |
|---------------------------------------|
                    |
                    |   StartAddSentencaBase(
                    |       ctx context.Context, 
                    |       sentenca []consts.ResponseAutosRow) 
                    |
        |---------------------------|
        |       Salva Registro      |
        |---------------------------|
                    |
                    |   salvaRegistro(
                    |       idPje, 
                    |       classe, 
                    |       assunto, 
                    |       natureza, 
                    |       item.Tipo, 
                    |       item.Tema, 
                    |       fonte, 
                    |       item.Paragrafos)
                    |
        |---------------------------|
        |   Gera Embbeding          |
        |---------------------------|
                    |
                    |   ialib.GetDocumentoEmbeddings(raw)
                    |
        |-----------------------------------|
        |   Salva em "rag_doc_embedding"    |
        |-----------------------------------|  
                    |
                    |   services.BaseServiceGlobal
                    |   .InserirDocumento(doc)
                    |
                            
                    
                                   
