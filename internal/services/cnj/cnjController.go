// cnjController.go
// Rotinas para consultas na API pública do CNJ
// Datas Revisão: 07/12/2024.

package cnj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ocrserver/api/handler/response"
	"ocrserver/internal/config"
	"ocrserver/internal/utils/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func BuscarProcessoCnj(numeroProcesso string) (*ResponseCnjPublicApi, error) {
	apiKey := config.CnjPublicApiKey
	cnjUrl := config.CnjPublicApiUrl

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

	if respostaCnj.Hits.Total.Value != 0 {
		log.Printf("Conteúdo devolvido")
		return &respostaCnj, nil
	}

	return nil, nil
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
func GetProcessoFromCnj(c *gin.Context) {
	//Generate request ID for tracing
	requestID := uuid.New().String()
	var requestData struct {
		NumeroProcesso string `json:"numeroProcesso"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		response.HandleError(c, http.StatusBadRequest, "Formato inválido", "", requestID)
		logger.Log.Error("JSON com Formato inválido", err.Error())
		return
	}

	if requestData.NumeroProcesso == "" {

		// response := msgs.CreateResponseMessage("Número do processo não indicado!")
		// c.JSON(http.StatusBadRequest, response)
		// return
		response.HandleError(c, http.StatusBadRequest, "Número do processo não indicado", "", requestID)
		logger.Log.Error("Número do processo não indicado")
		return
	}

	respostaCnj, err := BuscarProcessoCnj(requestData.NumeroProcesso)
	if err != nil {

		// response := msgs.CreateResponseMessage("Erro ao buscar processo na API do CNJ!")
		// c.JSON(http.StatusInternalServerError, response)
		// return
		response.HandleError(c, http.StatusBadRequest, "Erro ao buscar processo na API do CNJ!", "", requestID)
		logger.Log.Error("Erro ao buscar processo na API do CNJ!")
		return
	}

	if respostaCnj != nil {

		rsp := gin.H{
			"cnj": respostaCnj,
		}
		//c.JSON(http.StatusOK, response)
		c.JSON(http.StatusOK, response.NewSuccess(rsp, requestID))
		return

	} else {

		// rsp := gin.H{
		// 	"message": "Processo não localizado!",
		// }
		//c.JSON(http.StatusNoContent, response)
		c.JSON(http.StatusNotFound, response.NewError(http.StatusNotFound, "Processo não localizado!", "", requestID))
	}
}
