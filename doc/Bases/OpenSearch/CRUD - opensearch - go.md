// Indexa (cria/upsert) um contexto.
func (idx *ContextoIndexType) Indexa(
	nrProc string,
	juizo string,
	classe string,
	assunto string,
	usernameInc string,

) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}

	// ***** Criação do ID_CTXT  *************************
	idv7, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar uuidv7: %w", err)
	}
	idCtxt := idv7.String()
	//****************************************************
	now := time.Now()
	//*********************************
	body := ContextoRow{
		IdCtxt:           idCtxt,
		NrProc:           nrProc,
		Juizo:            juizo,
		Classe:           classe,
		Assunto:          assunto,
		PromptTokens:     0,
		CompletionTokens: 0,
		DtInc:            now,
		UsernameInc:      usernameInc,
		Status:           "S",
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Index(
		ctx,
		opensearchapi.IndexReq{
			Index:      idx.indexName,
			DocumentID: idCtxt, // Estou usando o id_ctxt como _id do documento
			Body:       opensearchutil.NewJSONReader(body),
			Params: opensearchapi.IndexParams{
				Refresh: "true",
			},
		})
	if err != nil {
		msg := fmt.Sprintf("Erro ao realizar indexação: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	return &ResponseContextoRow{
		Id:               idCtxt, // ✅ como você usou DocumentID=idCtxt, _id=idCtxt
		IdCtxt:           idCtxt,
		NrProc:           nrProc,
		Juizo:            juizo,
		Classe:           classe,
		Assunto:          assunto,
		PromptTokens:     0,
		CompletionTokens: 0,
		DtInc:            now,
		UsernameInc:      usernameInc,
		Status:           "S",
	}, nil
}

func (idx *ContextoIndexType) Update(
	idCtxt string,
	juizo string,
	classe string,
	assunto string,
) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	if strings.TrimSpace(idCtxt) == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}

	//ATENÇÃO: Não podemos usar a estrutura genérica do registro, salvo se todos os campos 
	//estiverem sendo alterados. Os campos não preenchidos importam no preenchimento do 
	//registro com valores zerados/vazios. 
	//Ou seja, todos os campos presentes em uma estrutura são considerados como campos a 
	//serem modificado no registro do OpenSearch, resultado em campos zerados se não fo-
	//rem passados valores.
	//Se o update é parcial, precisamos criar uma estrutura sob medida contendo apenas os
	//campos a alterar. Além disso, O json deve NESESSARIAMENTE conter o field "doc":
	//**
	//Exemplo: types.JsonMap{doc:types.JsonMap{fields}}
	
	body := types.JsonMap{
		"doc": types.JsonMap{
			"juizo":   juizo,
			"classe":  classe,
			"assunto": assunto,
		},
		"_source": true, // tenta devolver o source atualizado
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Update(
		ctx,
		opensearchapi.UpdateReq{
			Index:      idx.indexName,
			DocumentID: idCtxt, // ✅ se você adotou _id=id_ctxt
			Body:       opensearchutil.NewJSONReader(body),
			Params: opensearchapi.UpdateParams{
				Refresh: "true",
			},
		},
	)
	if err != nil {
		msg := fmt.Sprintf("Erro realizar update: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	//Pego o retorno do Update
	var result UpdateResponseGeneric[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	src := result.Get.Source
	//logger.Log.Infof("\nsrc.IdCtxt=%s", src.IdCtxt)
	// fallback mínimo caso não venha _source
	if src.IdCtxt == "" {
		src.IdCtxt = idCtxt
		src.Juizo = juizo
		src.Classe = classe
		src.Assunto = assunto
	}

	return &ResponseContextoRow{
		Id:               idCtxt,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}, nil
}

// DeleteByID deleta um documento diretamente pelo _id do OpenSearch
func (idx *ContextoIndexType) Delete(id string) error {
	if idx == nil || idx.osCli == nil {
		err := fmt.Errorf("OpenSearch não conectado")
		logger.Log.Error(err.Error())
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		err := fmt.Errorf("id vazio")
		logger.Log.Error(err.Error())
		return err
	}

	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	res, err := idx.osCli.Document.Delete(
		ctx,
		opensearchapi.DocumentDeleteReq{
			Index:      idx.indexName,
			DocumentID: id,
			Params: opensearchapi.DocumentDeleteParams{
				// ✅ Melhor opção para “sumir da lista” logo após o delete:
				Refresh: "true",
			},
		})

	if err != nil {
		msg := fmt.Sprintf("Erro realizar delete: %v", err)
		logger.Log.Error(msg)
		return err
	}
	if err = ReadOSErr(res.Inspect().Response); err != nil {
		return err
	}
	defer res.Inspect().Response.Body.Close()

	return nil
}

// Consulta por _id
func (idx *ContextoIndexType) ConsultaById(id string) (*ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("id vazio")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//Cria o objeto da requisição
	req := opensearchapi.DocumentGetReq{
		Index:      idx.indexName,
		DocumentID: id,
	}
	//Executa passando o objeto da requisição
	res, err := idx.osCli.Document.Get(
		ctx,
		req,
	)
	if err != nil {
		msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result DocumentGetResponse[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	if !result.Found {
		logger.Log.Infof("id=%s não encontrado (found=false)", id)
		return nil, nil
	}

	src := result.Source

	doc := ResponseContextoRow{
		Id:               result.ID,
		IdCtxt:           src.IdCtxt,
		NrProc:           src.NrProc,
		Juizo:            src.Juizo,
		Classe:           src.Classe,
		Assunto:          src.Assunto,
		PromptTokens:     src.PromptTokens,
		CompletionTokens: src.CompletionTokens,
		DtInc:            src.DtInc,
		UsernameInc:      src.UsernameInc,
		Status:           src.Status,
	}
	return &doc, nil
}

// Consultar documentos por id_ctxt
func (idx *ContextoIndexType) ConsultaByIdCtxt(idCtxt string) ([]ResponseContextoRow, error) {
	if idx == nil || idx.osCli == nil {
		return nil, fmt.Errorf("OpenSearch não conectado")
	}
	idCtxt = strings.TrimSpace(idCtxt)
	if idCtxt == "" {
		return nil, fmt.Errorf("idCtxt vazio")
	}
	ctx, cancel := NewCtx(idx.timeout)
	defer cancel()

	//----  Crio o body
	query := types.JsonMap{
		"size": QUERY_MAX_SIZE,
		"query": types.JsonMap{
			"term": types.JsonMap{
				"id_ctxt": idCtxt,
			},
		},
	}

	//Crio a SearchReq
	req := opensearchapi.SearchReq{
		Indices: []string{idx.indexName},
		Body:    opensearchutil.NewJSONReader(query),
	}

	//Executo a chamada da busca
	res, err := idx.osCli.Search(
		ctx,
		&req,
	)

	if err != nil {
		msg := fmt.Sprintf("Erro realizar consulta by query: %v", err)
		logger.Log.Error(msg)
		return nil, err
	}
	if err := ReadOSErr(res.Inspect().Response); err != nil {
		return nil, err
	}
	defer res.Inspect().Response.Body.Close()

	var result SearchResponseGeneric[ContextoRow]
	if err := json.NewDecoder(res.Inspect().Response.Body).Decode(&result); err != nil {
		logger.Log.Errorf("Erro ao decodificar JSON: %v", err)
		return nil, err
	}

	// ✅ Correção do panic
	if len(result.Hits.Hits) == 0 {
		return nil, nil
	}

	docs := make([]ResponseContextoRow, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		src := hit.Source
		docs = append(docs, ResponseContextoRow{

			Id:               hit.ID,
			IdCtxt:           src.IdCtxt,
			NrProc:           src.NrProc,
			Juizo:            src.Juizo,
			Classe:           src.Classe,
			Assunto:          src.Assunto,
			PromptTokens:     src.PromptTokens,
			CompletionTokens: src.CompletionTokens,
			DtInc:            src.DtInc,
			UsernameInc:      src.UsernameInc,
			Status:           src.Status,
		})
	}
	return docs, nil
}

