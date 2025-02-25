package elastic

import (
	"crypto/tls"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"log"
	"net"
	"net/http"
	"ocrserver/internal/config"
	"sync"
	"time"
)

type ElasticServerType struct {
	es *elasticsearch.Client
}

// Instância global para compartilhamento
var (
	EServer ElasticServerType
	once    sync.Once
)

// Função para criar um novo cliente Elasticsearch
func NewElasticServer() *ElasticServerType {
	es, err := EServer.GetClienteElastic()
	if err != nil {
		log.Println("Erro ao obter uma instância do cliente Elasticsearch:", err)
		return nil
	}

	return &ElasticServerType{es: es}
}

// Obtém as configurações do Elasticsearch a partir de variáveis de ambiente
func getConfigElasticServer() elasticsearch.Config {
	var esHost string

	// Verifica se o host e a porta foram configurados corretamente
	if config.ElasticHost == "" || config.ElasticPort == "" {
		esHost = "http://localhost:9200"
		log.Println("Aviso: Usando host padrão para Elasticsearch.")
	} else {
		esHost = config.ElasticHost + ":" + config.ElasticPort // Corrigida a concatenação correta
	}

	// Log para depuração
	log.Println("Conectando ao Elasticsearch em:", esHost)

	// Retorna a configuração do cliente Elasticsearch
	cfg := elasticsearch.Config{
		Addresses: []string{esHost},
		Username:  config.ElasticUser,
		Password:  config.ElasticPassword,
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

// Inicializa o cliente Elasticsearch
func InitializeElasticServer() error {
	var err error

	once.Do(func() {
		cfg := getConfigElasticServer()
		esCliente, errInit := elasticsearch.NewClient(cfg)
		if errInit != nil {
			log.Printf("Erro ao inicializar Elasticsearch: %v", errInit)
			err = errInit
			return
		}
		EServer.es = esCliente
		log.Println("Elasticsearch conectado com sucesso!")
	})
	return err
}

// Obtém a instância do cliente Elasticsearch
func (es *ElasticServerType) GetClienteElastic() (*elasticsearch.Client, error) {
	if es.es == nil {
		log.Println("Erro: Elasticsearch não conectado.")
		return nil, fmt.Errorf("erro ao tentar conectar ao Elasticsearch")
	}
	return es.es, nil
}

// Simula fechamento da conexão
func (es *ElasticServerType) CloseConn() {
	if es.es != nil {
		log.Println("Encerrando conexão com Elasticsearch (não há fechamento explícito necessário).")
	}
}
