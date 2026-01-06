package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
)

type SearchResponseGeneric[T any] struct {
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore *float32 `json:"max_score"`
		Hits     []struct {
			Index  string   `json:"_index"`
			ID     string   `json:"_id"`
			Score  *float64 `json:"_score,omitempty"`
			Source T        `json:"_source"`
			Sort   []any    `json:"sort,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

type DocumentGetResponse[T any] struct {
	Index       string `json:"_index"`
	ID          string `json:"_id"`
	Version     int    `json:"_version"`
	SeqNo       int    `json:"_seq_no"`
	PrimaryTerm int    `json:"_primary_term"`
	Found       bool   `json:"found"`
	Source      T      `json:"_source"`
}

// *********   HELPER  ********************

func NewCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), timeout)
}

func ReadOSErr(res *opensearch.Response) error {
	if res == nil {
		return fmt.Errorf("resposta nula do OpenSearch")
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return nil
	}

	b, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	res.Body = io.NopCloser(bytes.NewReader(b)) // permite re-leitura se alguém precisar

	return fmt.Errorf("opensearch status=%d: %s", res.StatusCode, string(b))
}

// decodeJSONHTTP lê o body uma única vez e decodifica em out (evita problemas de body consumido).
func DecodeJSONHTTP[T any](r *opensearch.Response, out *T) error {
	if r == nil {
		return fmt.Errorf("resposta HTTP nula")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if r.StatusCode < 200 || r.StatusCode >= 300 {
		return fmt.Errorf("opensearch status=%d: %s", r.StatusCode, strings.TrimSpace(string(body)))
	}
	if len(body) == 0 {
		return fmt.Errorf("body vazio")
	}
	return json.Unmarshal(body, out)
}
