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
	"strings"
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

	if !validarNumeroUnicoProcesso(requestData.NumeroProcesso) {

		logger.Log.Errorf("Número do processo não é válido %s", requestData.NumeroProcesso)
		response.HandleError(c, http.StatusBadRequest, "Número do processo não é válido", "", requestID)

		return
	}

	respostaCnj, err := obj.BuscarProcessoCnj(requestData.NumeroProcesso)
	if err != nil {
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar processo na API do CNJ!", "", requestID)
		logger.Log.Error("Erro ao buscar processo na API do CNJ!")
		return
	}

	rsp := gin.H{
		"metadados": respostaCnj,
		"message":   "Processo localizado com sucesso!",
	}
	response.HandleSuccess(c, http.StatusOK, rsp, requestID)
}

// validarNumeroUnicoProcesso verifica se o número está correto conforme CNJ (Res. 65/2008)
func validarNumeroUnicoProcesso(numero string) bool {
	// Remove pontos e traços
	numeroLimpo := strings.ReplaceAll(numero, ".", "")
	numeroLimpo = strings.ReplaceAll(numeroLimpo, "-", "")

	// Regex da estrutura NNNNNNNDD AAAA J TR OOOO
	re := regexp.MustCompile(`^(\d{7})(\d{2})(\d{4})(\d{1})(\d{2})(\d{4})$`)
	matches := re.FindStringSubmatch(numeroLimpo)
	if matches == nil {
		return false
	}

	nSeq := matches[1]
	dvInformado := matches[2]
	ano := matches[3]
	j := matches[4]
	tr := matches[5]
	origem := matches[6]

	// Recria o número base (sem DV)
	numeroBase := fmt.Sprintf("%s%s%s%s%s00", nSeq, ano, j, tr, origem)

	// Converte para inteiro de forma iterativa para evitar overflow
	mod := modulo97(numeroBase)

	dvCalculado := 98 - mod
	if dvCalculado < 10 {
		return fmt.Sprintf("0%d", dvCalculado) == dvInformado
	}
	return fmt.Sprintf("%d", dvCalculado) == dvInformado
}

// modulo97 implementa o cálculo iterativo ISO 7064 Mod 97-10
func modulo97(num string) int {
	const base = 97
	resto := 0

	for len(num) > 0 {
		tamanho := 9
		if len(num) < tamanho {
			tamanho = len(num)
		}
		parte := num[:tamanho]
		num = num[tamanho:]

		val, _ := strconv.Atoi(fmt.Sprintf("%d%s", resto, parte))
		resto = val % base
	}
	return resto
}
