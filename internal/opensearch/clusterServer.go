package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/v4" // ✅ Apenas versão v4
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"ocrserver/internal/config"
	"ocrserver/internal/utils/logger"
)

// Estrutura para o cliente OpenSearch
type ClusterServerType struct {
	client *opensearchapi.Client
	cfg    config.Config
}

// Instância global para compartilhamento
var (
	OpenSearchGlobal    ClusterServerType
	onceOpenserchGlobal sync.Once
)

// Inicializa o cliente OpenSearch
func InitOpenSearchService() error {
	var err error

	onceOpenserchGlobal.Do(func() {
		cfg := getConfigOpenSearchServer()

		config := opensearchapi.Config{Client: cfg}

		client, errInit := opensearchapi.NewClient(config)

		if errInit != nil {
			log.Printf("Erro ao inicializar OpenSearch: %v", errInit)
			err = errInit
			return
		}
		OpenSearchGlobal.client = client
		log.Println("OpenSearch conectado com sucesso!")
	})
	return err
}

// Função para criar um novo cliente OpenSearch
func NewClusterServer(cfg config.Config) *ClusterServerType {

	client, err := OpenSearchGlobal.GetClient()
	if err != nil {
		log.Println("Erro ao obter uma instância do cliente OpenSearch:", err)
		return nil
	}

	return &ClusterServerType{client: client, cfg: cfg}
}

// Obtém as configurações do OpenSearch a partir de variáveis de ambiente
func getConfigOpenSearchServer() opensearch.Config {
	var osHost string

	// Verifica se o host e a porta foram configurados corretamente
	if config.GlobalConfig.OpenSearchHost == "" || config.GlobalConfig.OpenSearchPort == "" {
		osHost = "http://localhost:9200"
		log.Println("Aviso: Usando host padrão para OpenSearch.")
	} else {
		osHost = config.GlobalConfig.OpenSearchHost + ":" + config.GlobalConfig.OpenSearchPort
	}

	// Log para depuração
	log.Println("Conectando ao OpenSearch em:", osHost)

	// Retorna a configuração do cliente OpenSearch
	cfg := opensearch.Config{
		Addresses: []string{osHost},
		Username:  config.GlobalConfig.OpenSearchUser,
		Password:  config.GlobalConfig.OpenSearchPassword,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 5 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}

	return cfg
}

// Obtém a instância do cliente OpenSearch
func (obj *ClusterServerType) GetClient() (*opensearchapi.Client, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	if obj.client == nil {
		log.Println("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao tentar conectar ao OpenSearch")
	}
	return obj.client, nil
}

// Simula fechamento da conexão
func (obj *ClusterServerType) CloseConn() {
	if obj.client != nil {
		log.Println("Encerrando conexão com OpenSearch (não há fechamento explícito necessário).")
	}
}

// Obter informações do cluster
func (obj *ClusterServerType) Info() (*opensearchapi.ClusterHealthResp, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return nil, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	res, err := obj.client.Cluster.Health(context.Background(), &opensearchapi.ClusterHealthReq{})
	if err != nil {
		log.Printf("Erro ao extrair informações do cluster OpenSearch: %v", err)
		return nil, err
	}
	return res, nil
}

// Verifica se o índice existe
func (obj *ClusterServerType) IndicesExists(indexStr string) (bool, error) {
	if obj == nil {
		logger.Log.Error("Tentativa de uso de serviço não iniciado.")
		return false, fmt.Errorf("tentativa de uso de serviço não iniciado")
	}
	res, err := obj.client.Indices.Exists(
		context.Background(),
		opensearchapi.IndicesExistsReq{
			Indices: []string{indexStr},
		})
	if err != nil {
		log.Printf("Erro ao verificar a existência do índice %s no OpenSearch: %v", indexStr, err)
		return false, err
	}
	defer res.Body.Close()

	return res.StatusCode == http.StatusOK, nil
}
