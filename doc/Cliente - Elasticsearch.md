res, err := cliente.esCli.Index(
		indexName,
		bytes.NewReader(data),
		//cliente.esCli.Index.WithDocumentID(""),  // Document ID
		cliente.esCli.Index.WithRefresh("true"), // Refresh
	)
	if err != nil {
		log.Printf("Erro ao indexar documento no Elasticsearch: %v", err)
		return nil, err
	}
	defer res.Body.Close()
