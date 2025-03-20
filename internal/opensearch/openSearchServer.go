package opensearch

import (
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
)

// Estrutura para o cliente OpenSearch
type OpenSearchServerType struct {
	client *opensearchapi.Client
}

// Instância global para compartilhamento
var (
	OServer OpenSearchServerType
	once    sync.Once
)

// Função para criar um novo cliente OpenSearch
func NewOpenSearchServer() *OpenSearchServerType {
	client, err := OServer.GetClient()
	if err != nil {
		log.Println("Erro ao obter uma instância do cliente OpenSearch:", err)
		return nil
	}

	return &OpenSearchServerType{client: client}
}

// Obtém as configurações do OpenSearch a partir de variáveis de ambiente
func getConfigOpenSearchServer() opensearch.Config {
	var osHost string

	// Verifica se o host e a porta foram configurados corretamente
	if config.OpenSearchHost == "" || config.OpenSearchPort == "" {
		osHost = "http://localhost:9200"
		log.Println("Aviso: Usando host padrão para OpenSearch.")
	} else {
		osHost = config.OpenSearchHost + ":" + config.OpenSearchPort
	}

	// Log para depuração
	log.Println("Conectando ao OpenSearch em:", osHost)

	// Retorna a configuração do cliente OpenSearch
	cfg := opensearch.Config{
		Addresses: []string{osHost},
		Username:  config.OpenSearchUser,
		Password:  config.OpenSearchPassword,
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

// Inicializa o cliente OpenSearch
func InitializeOpenSearchServer() error {
	var err error

	once.Do(func() {
		cfg := getConfigOpenSearchServer()

		config := opensearchapi.Config{Client: cfg}

		client, errInit := opensearchapi.NewClient(config)

		if errInit != nil {
			log.Printf("Erro ao inicializar OpenSearch: %v", errInit)
			err = errInit
			return
		}
		OServer.client = client
		log.Println("OpenSearch conectado com sucesso!")
	})
	return err
}

// Obtém a instância do cliente OpenSearch
func (os *OpenSearchServerType) GetClient() (*opensearchapi.Client, error) {
	if os.client == nil {
		log.Println("Erro: OpenSearch não conectado.")
		return nil, fmt.Errorf("erro ao tentar conectar ao OpenSearch")
	}
	return os.client, nil
}

// Simula fechamento da conexão
func (os *OpenSearchServerType) CloseConn() {
	if os.client != nil {
		log.Println("Encerrando conexão com OpenSearch (não há fechamento explícito necessário).")
	}
}
