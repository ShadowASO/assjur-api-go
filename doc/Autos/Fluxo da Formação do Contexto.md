		                Fluxo da Formação do Contexto

POST  "/contexto/documentos/autua"

	fonte: autosTempHandler.go
	
	AutuarDocumentosHandler(gin.Context)
	
		ProcessarDocumento(idCtxt, idDoc)
		
		    . extrai registro idDoc do índice "autos_temp"
		    
		    . pega o prompt de análise na tabela "prompts"
		    
		    . processa o texto usando o prompt para gerar o JSON
		    
		    . verifica se o documentos está duplicado
		    
		    . incluir o JSON no índice "autos"
		    
		    . gera o embedding do conteúdo do JSON
		    
		    . insere o embedding no índice "autos_json_embedding"
		
