/*
---------------------------------------------------------------------------------------
File: userService.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package opensearch

import (
	"fmt"
	//"ocrserver/internal/services"
	"ocrserver/internal/utils/logger"

	"sync"
)

type IndexServiceType struct {
	Model *IndexModelosType
}

var IndexService *IndexServiceType
var onceIndexService sync.Once

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitIndexService(Model *IndexModelosType) {
	onceIndexService.Do(func() {
		IndexService = &IndexServiceType{
			Model: Model,
		}

		logger.Log.Info("Global AutosService configurado com sucesso.")
	})
}

func NewIndexService(Model *IndexModelosType) *IndexServiceType {
	return &IndexServiceType{
		Model: Model,
	}
}

func (obj *IndexServiceType) GetDocumentoById(id string) (*ResponseModelos, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de utilizar CnjApi global sem inicializá-la.")
		return nil, fmt.Errorf("CnjApi global não configurada")
	}
	doc, err := obj.Model.ConsultaDocumentoById(id)
	if err != nil {
		logger.Log.Info("Erro ao selecionar documentos dos autos!")
		return nil, err
	}
	return doc, nil
}

/*
*
Obtem o embedding de cada campo texto do index Modelos e devolve uma strutura.
*/
// func (obj *IndexModelosType) GetDocumentoEmbeddings(doc ModelosText) (ModelosEmbedding, error) {
// 	ctx := context.Background()
// 	modelo := ModelosEmbedding{
// 		Natureza:     doc.Natureza,
// 		Ementa:       doc.Ementa,
// 		Inteiro_teor: doc.Inteiro_teor,
// 	}

// 	// Gera o embedding da ementa
// 	ementaResp, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, modelo.Ementa)
// 	if err != nil {
// 		return modelo, fmt.Errorf("erro ao gerar embedding da ementa: %w", err)
// 	}

// 	//Converte o embedding para float32
// 	vector32 := services.OpenaiServiceGlobal.Float64ToFloat32Slice(ementaResp)
// 	modelo.EmentaEmbedding = vector32

// 	// Gera o embedding do inteiro teor
// 	teorResp, err := services.OpenaiServiceGlobal.GetEmbeddingFromText(ctx, doc.Inteiro_teor)
// 	if err != nil {
// 		return modelo, fmt.Errorf("erro ao gerar embedding do inteiro teor: %w", err)
// 	}

// 	vector32 = services.OpenaiServiceGlobal.Float64ToFloat32Slice(teorResp)
// 	modelo.InteiroTeorEmbedding = vector32

// 	return modelo, nil
// }
