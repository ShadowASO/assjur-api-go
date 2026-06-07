package pipeline

import (
	"context"
	"encoding/json"
	"sync"

	"fmt"

	"ocrserver/internal/consts"
	"ocrserver/internal/models/opensearch"
	"ocrserver/internal/services"
	openaiservice "ocrserver/internal/services/openai"

	"ocrserver/internal/utils/mserror"
	"ocrserver/internal/utils/mslogger"

	"strings"
)

type RetrieverType struct {
}

const MAX_REGS_BY_TEMA_RAG = 2

func NewRetrieverType() *RetrieverType {
	return &RetrieverType{}
}

func (service *RetrieverType) RecuperaAutosProcesso(ctx context.Context, idCtxt string) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	if len(autos) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhum documento processual foi localizado nos autos: %v", err)
		return nil, err
	}
	mslogger.LoggerGlobal.Infof("Documentos do processo recuperados: %d", len(autos))

	return autos, nil
}

func (service *RetrieverType) RecuperaAutosProcessoAsMessages(ctx context.Context, idCtxt string) ([]openaiservice.MessageResponseItem, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	messages := openaiservice.MsgGpt{}
	if len(autos) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhum documento processual foi localizado nos autos: %v", err)
		return messages.Messages, nil
	}

	for _, msg := range autos {
		messages.AddMessage(openaiservice.MessageResponseItem{
			Id:   "",
			Role: "user",
			Text: msg.DocJsonRaw,
		})
	}

	return messages.Messages, nil
}

/*
Recupera as sentenças judiciais proferidas nos autos do processo.
*/
func (service *RetrieverType) RecuperaAutosSentenca(ctx context.Context, idCtxt string) ([]consts.ResponseAutosRow, error) {

	autos, err := services.AutosServiceGlobal.GetAutosByContexto(idCtxt)

	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar os autos: %v", err)
		return nil, err
	}
	if len(autos) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhuma análise processual foi localizada: %v", err)
		return nil, err
	}
	//Procuro todos os registros com a natureza RAG_RESPONSE_ANALISE
	sentencas := []consts.ResponseAutosRow{}
	for _, row := range autos {
		if row.IdNatu == consts.NATU_DOC_SENTENCA {
			sentencas = append(sentencas, row)
		}
	}

	return sentencas, nil
}

func (service *RetrieverType) RecuperaPreAnaliseJuridica(
	ctx context.Context,
	idCtxt string,
) ([]opensearch.ResponseEventosRow, error) {

	eventos, err := services.EventosServiceGlobal.GetEventosByContexto(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao recuperar autos do contexto: %v", idCtxt, err)
		return nil, fmt.Errorf("erro ao recuperar autos do contexto: %w", err)
	}

	if len(eventos) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhum registro de autos encontrado para o contexto.", idCtxt)
		return nil, nil
	}

	documentos := make([]opensearch.ResponseEventosRow, 0)
	for _, row := range eventos {
		if row.IdNatu == consts.NATU_DOC_IA_PREANALISE {
			if strings.TrimSpace(row.DocJsonRaw) == "" {
				mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Pré-análise encontrada (id=%s) mas JSON está vazio.", idCtxt, row.Id)
				continue
			}
			documentos = append(documentos, row)
		}
	}

	if len(documentos) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhuma pré-análise válida (com JSON) encontrada entre %d autos.", idCtxt, len(eventos))
		return nil, nil
	}

	mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Recuperadas %d pré-análises válidas.", idCtxt, len(documentos))
	return documentos, nil
}

/*
Devolve todos os registros de Análise Jurídica realizadas pelo modelo de IA
*/
func (service *RetrieverType) RecuperaAnaliseJuridica(
	ctx context.Context,
	idCtxt string,
) ([]opensearch.ResponseEventosRow, error) {

	eventos, err := services.EventosServiceGlobal.GetEventosByContexto(idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("[id_ctxt=%s] Erro ao recuperar autos do contexto: %v", idCtxt, err)
		return nil, fmt.Errorf("erro ao recuperar autos do contexto: %w", err)
	}

	if len(eventos) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhum registro encontrado nos autos para o contexto.", idCtxt)
		return nil, nil
	}

	documentos := make([]opensearch.ResponseEventosRow, 0)
	for _, row := range eventos {
		if row.IdNatu == consts.NATU_DOC_IA_ANALISE {
			if strings.TrimSpace(row.DocJsonRaw) == "" {
				mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] análise encontrada (id=%s) mas JSON está vazio.", idCtxt, row.Id)
				continue
			}
			documentos = append(documentos, row)
		}
	}

	if len(documentos) == 0 {
		mslogger.LoggerGlobal.Warnf("[id_ctxt=%s] Nenhuma análise jurídica válida (com JSON) encontrada entre %d registros nos autos.", idCtxt, len(eventos))
		return nil, nil
	}

	mslogger.LoggerGlobal.Infof("[id_ctxt=%s] Recuperadas %d análises jurídicas válidas.", idCtxt, len(documentos))
	return documentos, nil
}

func (service *RetrieverType) RecuperaDoutrinaRAG_(ctx context.Context, idCtxt string) ([]opensearch.ResponseModelos, error) {

	//***   Recupera pré-análise
	preAnalise, err := service.RecuperaPreAnaliseJuridica(ctx, idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar busca de pré-análise: %v", err)
		return nil, mserror.NewError("Erro ao buscar pré-analise %s", err.Error())
	}

	if len(preAnalise) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhuma doutrina recuperada")
		return nil, nil
	}

	// Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, preAnalise[0].Doc)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, mserror.NewError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_DOUTRINA))
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar modelos de doutrina: %v", err)
		return nil, mserror.NewError("Erro ao consultar modelos de doutrina: %s", err.Error())
	}
	if len(docs) == 0 {
		mslogger.LoggerGlobal.Info("Nenhum modelo de doutrina retornado")
		return nil, nil
	}

	mslogger.LoggerGlobal.Infof("Documentos do doutrina recuperados: %d", len(docs))

	return docs, nil
}
func (service *RetrieverType) RecuperaAcordaoRAG(ctx context.Context, idCtxt string) ([]opensearch.ResponseModelos, error) {

	analise, err := service.RecuperaAnaliseJuridica(ctx, idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar acórdãos: %v", err)
		return nil, mserror.NewError("Erro ao recuperar acórdãos: %s", err.Error())
	}
	if len(analise) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhum acórdão localizado")
		return nil, nil
	}

	//Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, analise[0].Doc)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, mserror.NewError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_ACORDAO))
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar modelos de acórdão: %v", err)
		return nil, mserror.NewError("Erro ao consultar modelos de acórdão: %s", err.Error())
	}
	if len(docs) == 0 {
		mslogger.LoggerGlobal.Info("Nenhum modelo de acórdão retornado")
		return nil, nil
	}

	return docs, nil
}

func (service *RetrieverType) RecuperaSumulaRAG(ctx context.Context, idCtxt string) ([]opensearch.ResponseModelos, error) {

	analise, err := service.RecuperaAnaliseJuridica(ctx, idCtxt)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao recuperar súmulas: %v", err)
		return nil, mserror.NewError("Erro ao recuperar súmulas: %s", err.Error())
	}
	if len(analise) == 0 {
		mslogger.LoggerGlobal.Errorf("Nenhuma súmula recuperada")
		return nil, nil
	}

	//Converte a string de busca num embedding
	vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, analise[0].Doc)
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao gerar embeddings: %v", err)
		return nil, mserror.NewError("Erro ao gerar embedding: %s", err.Error())
	}

	docs, err := opensearch.ModelosServiceGlobal.ConsultaSemantica(vec32, opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_SUMULA))
	if err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao consultar modelos de súmula: %v", err)
		return nil, mserror.NewError("Erro ao consultar modelos de súmula: %s", err.Error())
	}
	if len(docs) == 0 {
		mslogger.LoggerGlobal.Info("Nenhum modelo de súmula retornado")
		return nil, nil
	}

	return docs, nil
}

// RecuperaBaseConhecimentos executa buscas semânticas concorrentes controladas
// para cada tema jurídico do campo RAG identificado durante a análise jurídica
// pelo modelo de IA. O campo  "DocJsonRaw" possui o objeto JSON gerado e con-
// verte para um objeto Go.
// Usa semáforo para limitar goroutines simultâneas e realiza deduplicação global ao final.
func (service *RetrieverType) RecuperaBaseConhecimentos(
	ctx context.Context,
	idCtxt string,
	analise opensearch.ResponseEventosRow) ([]opensearch.ResponseBaseRow, error) {
	mslogger.LoggerGlobal.Infof("Iniciando recuperação da Base de conhecimentos=%s", idCtxt)

	// 2️⃣ Converte o JSON armazenado em objeto Go
	var objAnalise AnaliseJuridicaIA
	docJson := analise.DocJsonRaw
	if err := json.Unmarshal([]byte(docJson), &objAnalise); err != nil {
		mslogger.LoggerGlobal.Errorf("Erro ao realizar unmarshal da análise: %v", err)
		return nil, mserror.NewError("Erro ao interpretar resposta da análise")
	}

	if len(objAnalise.Rag) == 0 {
		mslogger.LoggerGlobal.Warnf("Nenhuma questão jurídica encontrado na análise jurídica do processo %s", idCtxt)
		return nil, nil
	}

	// 3️⃣ Configuração de concorrência
	maxConcurrent := 10 // limite de goroutines simultâneas
	sema := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	resultsChan := make(chan []opensearch.ResponseBaseRow, len(objAnalise.Rag))

	// 4️⃣ Loop concorrente sobre os temas RAG
	for _, itemRag := range objAnalise.Rag {
		item := itemRag // captura da variável no escopo da goroutine

		wg.Add(1)
		go func() {
			defer wg.Done()
			sema <- struct{}{}        // ocupa um slot
			defer func() { <-sema }() // libera ao terminar

			queryText := strings.TrimSpace(fmt.Sprintf("%s: %s", item.Tema, item.Descricao))
			if queryText == "" {
				return
			}

			// 🔹 Gera embedding do texto do tema
			vec32, _, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, queryText)
			if err != nil {
				mslogger.LoggerGlobal.Errorf("Erro ao gerar embedding RAG (%s): %v", item.Tema, err)
				return
			}

			// 🔹 Executa consulta semântica no índice base_doc_embedding
			docs, err := opensearch.BaseIndexGlobal.ConsultaSemantica(
				vec32,
				//opensearch.GetNaturezaModelo(opensearch.MODELO_NATUREZA_SENTENCA),
				"",
			)
			if err != nil {
				mslogger.LoggerGlobal.Errorf("Erro ao consultar base RAG (%s): %v", item.Tema, err)
				return
			}

			if len(docs) == 0 {
				mslogger.LoggerGlobal.Infof("Nenhum documento retornado para tema '%s'", item.Tema)
				return
			}

			// 🔹 Mantém até n primeiros resultados por tema
			limite := MAX_REGS_BY_TEMA_RAG
			if len(docs) < limite {
				limite = len(docs)
			}

			resultsChan <- docs[:limite]
			mslogger.LoggerGlobal.Infof("Tema '%s' → %d documentos enviados ao canal", item.Tema, limite)
		}()
	}

	// 5️⃣ Goroutine de agregação: aguarda o fim de todas as buscas
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 6️⃣ Agrega todos os resultados brutos
	var resultadosBrutos []opensearch.ResponseBaseRow
	for docs := range resultsChan {
		resultadosBrutos = append(resultadosBrutos, docs...)
	}

	if len(resultadosBrutos) == 0 {
		mslogger.LoggerGlobal.Warn("Nenhum resultado bruto RAG retornado após execução concorrente")
		return nil, nil
	}

	// 7️⃣ Deduplicação global
	idsVistos := make(map[string]bool)
	resultadosUnicos := make([]opensearch.ResponseBaseRow, 0, len(resultadosBrutos))

	for _, doc := range resultadosBrutos {
		if idsVistos[doc.Id] {
			continue
		}
		idsVistos[doc.Id] = true
		resultadosUnicos = append(resultadosUnicos, doc)
	}

	// 8️⃣ Retorno final
	if len(resultadosUnicos) == 0 {
		mslogger.LoggerGlobal.Warn("Todos os resultados eram duplicados — vetor final vazio")
		return nil, nil
	}

	mslogger.LoggerGlobal.Infof("Busca RAG concorrente concluída: %d únicos (de %d brutos)",
		len(resultadosUnicos), len(resultadosBrutos))

	return resultadosUnicos, nil
}
