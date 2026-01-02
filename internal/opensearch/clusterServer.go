package opensearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"ocrserver/internal/config"
	"ocrserver/internal/utils/logger"
)

// Estrutura para o cliente OpenSearch
type ClusterServerType struct {
	client *opensearchapi.Client // ✅ ESTE é o client que tem Info(ctx, nil) e respostas com Inspect()
	cfg    config.Config
}

var (
	OpenSearchGlobal    ClusterServerType
	onceOpenserchGlobal sync.Once
)

func InitOpenSearchService() error {
	var errOut error

	onceOpenserchGlobal.Do(func() {
		osCfg := getConfigOpenSearchServer()

		// ✅ opensearchapi.Client é construído em cima do opensearch.Config (client “core”)
		client, err := opensearchapi.NewClient(opensearchapi.Config{
			//client, err := opensearch.NewClient(opensearch.Config{
			Client: osCfg,
		})
		if err != nil {
			errOut = fmt.Errorf("erro ao inicializar OpenSearch: %w", err)
			return
		}

		OpenSearchGlobal.client = client

		// smoke test (opcional)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := client.Info(ctx, nil)
		if err != nil {
			errOut = fmt.Errorf("opensearch info falhou: %w", err)
			return
		}
		defer res.Inspect().Response.Body.Close()

		if sc := res.Inspect().Response.StatusCode; sc < 200 || sc >= 300 {
			b, _ := io.ReadAll(res.Inspect().Response.Body)
			errOut = fmt.Errorf("opensearch info status=%d: %s", sc, strings.TrimSpace(string(b)))
			return
		}
	})

	return errOut
}

func NewClusterServer(cfg config.Config) *ClusterServerType {
	client, err := OpenSearchGlobal.GetClient()
	if err != nil {
		logger.Log.Errorf("Erro ao obter cliente OpenSearch: %v", err)
		return nil
	}
	return &ClusterServerType{client: client, cfg: cfg}
}

func getConfigOpenSearchServer() opensearch.Config {
	host := strings.TrimSpace(config.GlobalConfig.OpenSearchHost)
	port := strings.TrimSpace(config.GlobalConfig.OpenSearchPort)

	addr := "http://localhost:9200"
	if host != "" && port != "" {
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "http://" + host
		}
		host = strings.TrimRight(host, "/")
		addr = fmt.Sprintf("%s:%s", host, port)
	}

	return opensearch.Config{
		Addresses: []string{addr},
		Username:  config.GlobalConfig.OpenSearchUser,
		Password:  config.GlobalConfig.OpenSearchPassword,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 10 * time.Second,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
}

// func (obj *ClusterServerType) GetClient() (*opensearchapi.Client, error) {
func (obj *ClusterServerType) GetClient() (*opensearchapi.Client, error) {
	if obj == nil {
		return nil, fmt.Errorf("serviço OpenSearch não iniciado")
	}
	if obj.client == nil {
		return nil, fmt.Errorf("OpenSearch não conectado (client nil)")
	}
	return obj.client, nil
}

func (obj *ClusterServerType) Info(ctx context.Context) (int, string, error) {
	if obj == nil || obj.client == nil {
		return 0, "", fmt.Errorf("OpenSearch não conectado")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	res, err := obj.client.Info(ctx, nil) // ✅ v4 opensearchapi.Client
	if err != nil {
		return 0, "", err
	}
	defer res.Inspect().Response.Body.Close()

	sc := res.Inspect().Response.StatusCode
	b, _ := io.ReadAll(res.Inspect().Response.Body)
	return sc, string(b), nil
}

func (obj *ClusterServerType) IndicesExists(ctx context.Context, indexStr string) (bool, error) {
	if obj == nil || obj.client == nil {
		return false, fmt.Errorf("OpenSearch não conectado")
	}
	indexStr = strings.TrimSpace(indexStr)
	if indexStr == "" {
		return false, fmt.Errorf("index vazio")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	res, err := obj.client.Indices.Exists(ctx, opensearchapi.IndicesExistsReq{
		Indices: []string{indexStr},
	})
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		body, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("indices exists status=%d: %s",
			res.StatusCode, strings.TrimSpace(string(body)))
	}
}
