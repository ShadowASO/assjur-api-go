┌───────────────────────────────────────┐
│        Início do Pipeline             │
│        Análise do Processo            │
└───────────────────────────────────────┘
                    │
                    │   pipelineAnaliseProcesso(
                    │       ctx context.Context,
                    │       id_ctxt int,
                    │       msgs ialib.MsgGpt,
                    │       prevID string,
                    │   )
                    │
                    ▼
┌───────────────────────────────────────┐
│   Recupera os Documentos dos Autos    │
└───────────────────────────────────────┘
                    │
                    │   retriObj.RecuperaAutosProcesso(ctx, id_ctxt)
                    │
                    ▼
┌───────────────────────────────────────┐
│          Recupera Pré-Análise         │
└───────────────────────────────────────┘
                    │
                    │   retriObj.RecuperaPreAnaliseJudicial(ctx, id_ctxt)
                    │
                    │   if (existe pré-análise)
                    │
                    ▼
┌───────────────────────────────────────┐
│   Recupera Base de Conhecimento RAG   │
└───────────────────────────────────────┘
                    │
                    │   retriObj.RecuperaDoutrinaRAG(ctx, id_ctxt)
                    │
                    ▼
┌───────────────────────────────────────┐
│      Executa a Análise do Processo    │
└───────────────────────────────────────┘
                    │
                    │   genObj.ExecutaAnaliseProcesso(
                    │       ctx,
                    │       id_ctxt,
                    │       msgs,
                    │       prevID,
                    │       autos,
                    │       ragDoutrina,
                    │   )
                    │
                    │──────────────────────►
                    │                      ┌───────────────────────────────┐
                    │                      │      Seleciona um Prompt      │
                    │                      │       PROMPT_RAG_ANALISE       │
                    │                      └───────────────────────────────┘
                    │                                  │
                    │                                  │
                    │                                  │   services.PromptServiceGlobal.
                    │                                  │       GetPromptByNatureza(
                    │                                  │           consts.PROMPT_RAG_ANALISE,
                    │                                  │       )
                    │                                  │
                    │                      ┌───────────────────────────────┐
                    │                      │     Submete ao Modelo IA      │
                    │                      └───────────────────────────────┘
                    │                                  │
                    │                                  │   services.OpenaiServiceGlobal.
                    │                                  │       SubmitPromptResponse(
                    │                                  │           ctx,
                    │                                  │           messages,
                    │                                  │           prevID,
                    │                                  │           config.GlobalConfig.
                    │                                  │               OpenOptionModel,
                    │                                  │           ialib.REASONING_LOW,
                    │                                  │           ialib.VERBOSITY_LOW,
                    │                                  │       )
                    │                                  │
                    │                      ┌───────────────────────────────┐
                    │                      │  Atualiza Consumo de Tokens   │
                    │                      └───────────────────────────────┘
                    │                                  │
                    ◄───────────────────────────────────┘
                    │
                    ▼
┌───────────────────────────────────────┐
│     Salva o Resultado da Análise      │
└───────────────────────────────────────┘
                    │
                    │   service.salvarAnaliseProcesso(
                    │       ctx,
                    │       id_ctxt,
                    │       natuAnalise,
                    │       resposta,
                    │       "",
                    │   )
                    │
                    ▼
                (Fim do Pipeline)

