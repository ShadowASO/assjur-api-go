// cnjController.go
// Rotinas para consultas na API pública do CNJ
// Datas Revisão: 07/12/2024.

package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"

	"ocrserver/internal/config"
	"ocrserver/internal/handlers/response"
	"ocrserver/internal/utils/logger"
	"ocrserver/internal/utils/middleware"

	"github.com/gin-gonic/gin"
)

type ResponseCnjPublicApi struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				NumeroProcesso string `json:"numeroProcesso"`
				Classe         struct {
					Codigo int    `json:"codigo"`
					Nome   string `json:"nome"`
				} `json:"classe"`
				Sistema struct {
					Codigo int    `json:"codigo"`
					Nome   string `json:"nome"`
				} `json:"sistema"`
				Formato struct {
					Codigo int    `json:"codigo"`
					Nome   string `json:"nome"`
				} `json:"formato"`
				Tribunal                  string `json:"tribunal"`
				DataHoraUltimaAtualizacao string `json:"dataHoraUltimaAtualizacao"`
				Grau                      string `json:"grau"`
				Timestamp                 string `json:"@timestamp"`
				DataAjuizamento           string `json:"dataAjuizamento"`
				Movimentos                []struct {
					ComplementosTabelados []struct {
						Codigo    int    `json:"codigo"`
						Valor     int    `json:"valor"`
						Nome      string `json:"nome"`
						Descricao string `json:"descricao"`
					} `json:"complementosTabelados,omitempty"`
					Codigo   int    `json:"codigo"`
					Nome     string `json:"nome"`
					DataHora string `json:"dataHora"`
				} `json:"movimentos"`
				ID            string `json:"id"`
				NivelSigilo   int    `json:"nivelSigilo"`
				OrgaoJulgador struct {
					CodigoMunicipioIBGE int    `json:"codigoMunicipioIBGE"`
					Codigo              int    `json:"codigo"`
					Nome                string `json:"nome"`
				} `json:"orgaoJulgador"`
				Assuntos []struct {
					Codigo int    `json:"codigo"`
					Nome   string `json:"nome"`
				} `json:"assuntos"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type RequestCnjPublicApi struct {
	Query struct {
		Match struct {
			NumeroProcesso string `json:"numeroProcesso"`
		} `json:"match"`
	} `json:"query"`
}

const (
	HTTPStatusOK         = 200
	HTTPStatusNotFound   = 404
	HTTPStatusBadRequest = 400
)

type CnjServiceType struct {
	cfg *config.Config
}

var CnjApi *CnjServiceType
var onceInitCnjApi sync.Once

func NewCnjService(cfg *config.Config) *CnjServiceType {
	return &CnjServiceType{
		cfg: cfg,
	}
}

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitCnjGlobal(cfg *config.Config) {
	onceInitCnjApi.Do(func() {

		CnjApi = &CnjServiceType{
			cfg: cfg,
		}

		logger.Log.Info("CnjApi configurada com sucesso.")
	})
}

func (obj *CnjServiceType) BuscarProcessoCnj(numeroProcesso string) (*ResponseCnjPublicApi, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	apiKey := obj.cfg.CnjPublicApiKey
	cnjUrl := obj.cfg.CnjPublicApiUrl

	if apiKey == "" {
		return nil, fmt.Errorf("API key não configurada")
	}
	if cnjUrl == "" {
		return nil, fmt.Errorf("URL do CNJ não configurada")
	}

	processo := RequestCnjPublicApi{}
	processo.Query.Match.NumeroProcesso = numeroProcesso

	requestBody, err := json.Marshal(processo)
	if err != nil {
		fmt.Printf("Erro ao serializar a estrutura: %v\n", err)
	}

	req, err := http.NewRequest("POST", cnjUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("APIKey %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("falha na requisição: %s", resp.Status)
	}

	var respostaCnj ResponseCnjPublicApi
	if err := json.NewDecoder(resp.Body).Decode(&respostaCnj); err != nil {
		log.Printf("falha na requisição: %s", resp.Status)
		return nil, err
	}

	if respostaCnj.Hits.Total.Value == 0 {
		log.Printf("Nenhum valor devolvido pela API Pública!")
		return &respostaCnj, nil
	}

	return &respostaCnj, nil
}

/*
 * Verifica a existência dos metadados do processo no CNJ e os devolve na resposta
 *
 * - **Rota**: "/cnj/processo"
 * - **Params**:
 * - **Método**: POST
 * - **Status**: 200/400/
 * - **Body:
 *		{
 *		  "NumeroProcesso": "30021564620238060167"
 * 		}
 * - **Resposta**:
 *  	{
 *			"cnj":        respostaCnj,
 *		}
 */
func (obj *CnjServiceType) GetProcessoFromCnj(c *gin.Context) {
	//Generate request ID for tracing
	//requestID := uuid.New().String()
	requestID := middleware.GetRequestID(c)

	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		response.HandleError(c, http.StatusBadRequest, "Erro interno", "", requestID)
		return
	}
	var requestData struct {
		NumeroProcesso string `json:"numeroProcesso"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		logger.Log.Error("JSON com Formato inválido", err.Error())
		return
	}

	if requestData.NumeroProcesso == "" {

		response.HandleError(c, http.StatusBadRequest, "Número do processo não indicado", "", requestID)
		logger.Log.Error("Número do processo não indicado")
		return
	}

	if !ValidarNumeroUnicoProcesso(requestData.NumeroProcesso) {

		logger.Log.Error("Número do processo não é válido")
		response.HandleError(c, http.StatusBadRequest, "Número do processo não é válido", "", requestID)

		return
	}

	respostaCnj, err := obj.BuscarProcessoCnj(requestData.NumeroProcesso)
	if err != nil {
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar processo na API do CNJ!", "", requestID)
		logger.Log.Error("Erro ao buscar processo na API do CNJ!")
		return
	}

	// if respostaCnj == nil {

	// 	rsp := gin.H{
	// 		"cnj": respostaCnj,
	// 	}
	// 	//c.JSON(http.StatusOK, response)
	// 	c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
	// 	return

	// } else {

	// 	// rsp := gin.H{
	// 	// 	"message": "Processo não localizado!",
	// 	// }
	// 	//c.JSON(http.StatusNoContent, response)
	// 	c.JSON(http.StatusNotFound, response.NewError(http.StatusNotFound, "Processo não localizado!", "", requestID))
	// }
	rsp := gin.H{
		"metadados": respostaCnj,
		"message":   "Processo localizado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// ValidarNumeroUnicoProcesso valida a numeração CNJ de um processo judicial.
// Retorna true se o número for válido conforme o cálculo do dígito verificador (módulo 97).
func ValidarNumeroUnicoProcesso(numero string) bool {
	// 🔹 Remove pontos e traços
	re := regexp.MustCompile(`[.\-]`)
	numeroProcesso := re.ReplaceAllString(numero, "")

	// 🔹 Verificações básicas
	if len(numeroProcesso) < 14 {
		return false
	}
	if _, err := strconv.Atoi(numeroProcesso); err != nil {
		return false
	}

	// 🔹 Extrai partes do número CNJ
	digitoVerificadorExtraido, _ := strconv.Atoi(numeroProcesso[len(numeroProcesso)-13 : len(numeroProcesso)-11])
	vara := numeroProcesso[len(numeroProcesso)-4:]                            // 4 últimos dígitos
	tribunal := numeroProcesso[len(numeroProcesso)-6 : len(numeroProcesso)-4] // penúltimos 2
	ramo := numeroProcesso[len(numeroProcesso)-7 : len(numeroProcesso)-6]     // 1 dígito antes do tribunal
	anoInicio := numeroProcesso[len(numeroProcesso)-11 : len(numeroProcesso)-7]
	numeroSequencial := numeroProcesso[:len(numeroProcesso)-13]

	// 🔹 Preenche à esquerda com zeros até 7 dígitos
	if len(numeroSequencial) < 7 {
		numeroSequencial = fmt.Sprintf("%07s", numeroSequencial)
	}

	// 🔹 Calcula o dígito verificador conforme módulo 97
	valor := numeroSequencial + anoInicio + ramo + tribunal + vara + "00"
	mod := bcmod(valor, 97)
	digitoVerificadorCalculado := 98 - mod

	return digitoVerificadorExtraido == digitoVerificadorCalculado
}

// bcmod implementa o cálculo de módulo 97 sobre uma string numérica longa,
// similar ao comportamento do algoritmo em JavaScript.
func bcmod(x string, y int) int {
	mod := 0
	for len(x) > 0 {
		take := 5
		if len(x) < take {
			take = len(x)
		}

		chunk := x[:take]
		x = x[take:]

		num, _ := strconv.Atoi(fmt.Sprintf("%d%s", mod, chunk))
		mod = num % y
	}
	return mod
}
