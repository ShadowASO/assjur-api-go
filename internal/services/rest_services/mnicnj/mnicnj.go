/*
---------------------------------------------------------------------------------------
File: mni.go
Autor: Aldenor
Data: 05-06-2026
Alteração: 05-06-2026
---------------------------------------------------------------------------------------
Cliente REST específico para comunicação com o microsserviço MNI.

Este cliente utiliza a biblioteca ClienteHTTP para realizar chamadas HTTP ao serviço
responsável por consultar o PJe/MNI.

Uso típico:

	mniClient, err := mni.New(msclientehttp.ConfigClienteHTTP{
		Name:               cfg.MniName,
		BaseURL:            cfg.MniServiceURL,
		Timeout:            90 * time.Second,
		Debug:              cfg.MniClientDebug,
		InsecureSkipVerify: cfg.MniInsecureSkipVerify,
	})
	if err != nil {
		panic(err)
	}

	resp, err := mniClient.ConsultarDocumento(ctx, mni.ConsultaProcessoPJERequest{
		NumeroProcesso: "3003017-61.2025.8.06.0167",
		UsuarioCpf:     "00000000000",
		UsuarioSenha:   "senha",
		Documentos:     []string{"157207550"},
		Formato:        "texto",
	})
---------------------------------------------------------------------------------------
*/

package mnicnj

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"ocrserver/internal/pkg/msclientehttp"
)

const (
	defaultServiceName = "mni-srv"
	defaultTimeout     = 90 * time.Second

	routeConsultaCabecalho  = "/mni/processo/cabecalho"
	routeConsultaDocumentos = "/mni/processo/documentos"
	routeConsultaDocumento  = "/mni/processo/documento"
	routeConsultaMovimentos = "/mni/processo/movimentos"
	routeConsultaCompleta   = "/mni/processo/completo"
)

const (
	FormatoEntregaAuto      = "auto"
	FormatoEntregaHTML      = "html"
	FormatoEntregaTexto     = "texto"
	FormatoEntregaBase64    = "base64"
	FormatoEntregaMetadados = "metadados"
)

type ClientMNI struct {
	http *msclientehttp.ClienteHTTP
}

func (c *ClientMNI) validate() error {
	if c == nil || c.http == nil {
		return errors.New("mni client não inicializado")
	}

	return nil
}

func New(cfg msclientehttp.ConfigClienteHTTP) (*ClientMNI, error) {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = defaultServiceName
	}

	httpClient, err := msclientehttp.New(msclientehttp.ConfigClienteHTTP{
		Name:               name,
		BaseURL:            cfg.BaseURL,
		Timeout:            timeout,
		Debug:              cfg.Debug,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	})
	if err != nil {
		return nil, err
	}

	return &ClientMNI{
		http: httpClient,
	}, nil
}

func MustNew(cfg msclientehttp.ConfigClienteHTTP) *ClientMNI {
	client, err := New(cfg)
	if err != nil {
		panic(err)
	}

	return client
}

func NewFromURL(baseURL string) (*ClientMNI, error) {
	return New(msclientehttp.ConfigClienteHTTP{
		Name:               defaultServiceName,
		BaseURL:            baseURL,
		Timeout:            defaultTimeout,
		Debug:              false,
		InsecureSkipVerify: false,
	})
}

func (c *ClientMNI) ConsultarCabecalho(
	ctx context.Context,
	req ConsultaProcessoPJERequest,
) (*ConsultaProcessoPJEResponse, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := validarRequisicaoBase(req); err != nil {
		return nil, err
	}

	var resp APIResponse[*ConsultaProcessoPJEResponse]

	if err := c.http.PostJSON(ctx, routeConsultaCabecalho, nil, req, &resp); err != nil {
		return nil, err
	}

	return parseAPIResponse(resp, "erro ao consultar cabeçalho do processo")
}

func (c *ClientMNI) ConsultarListaDocumentos(
	ctx context.Context,
	req ConsultaProcessoPJERequest,
) (*ConsultaProcessoPJEResponse, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := validarRequisicaoBase(req); err != nil {
		return nil, err
	}

	var resp APIResponse[*ConsultaProcessoPJEResponse]

	if err := c.http.PostJSON(ctx, routeConsultaDocumentos, nil, req, &resp); err != nil {
		return nil, err
	}

	return parseAPIResponse(resp, "erro ao consultar lista de documentos do processo")
}

func (c *ClientMNI) ConsultarMovimentos(
	ctx context.Context,
	req ConsultaProcessoPJERequest,
) (*ConsultaProcessoPJEResponse, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := validarRequisicaoBase(req); err != nil {
		return nil, err
	}

	var resp APIResponse[*ConsultaProcessoPJEResponse]

	if err := c.http.PostJSON(ctx, routeConsultaMovimentos, nil, req, &resp); err != nil {
		return nil, err
	}

	return parseAPIResponse(resp, "erro ao consultar movimentos do processo")
}

func (c *ClientMNI) ConsultarCompleto(
	ctx context.Context,
	req ConsultaProcessoPJERequest,
) (*ConsultaProcessoPJEResponse, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := validarRequisicaoBase(req); err != nil {
		return nil, err
	}

	var resp APIResponse[*ConsultaProcessoPJEResponse]

	if err := c.http.PostJSON(ctx, routeConsultaCompleta, nil, req, &resp); err != nil {
		return nil, err
	}

	return parseAPIResponse(resp, "erro ao consultar processo completo")
}

func (c *ClientMNI) ConsultarDocumento(
	ctx context.Context,
	req ConsultaProcessoPJERequest,
) (*ConsultaDocumentoPJEResponse, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	if err := validarRequisicaoDocumento(req); err != nil {
		return nil, err
	}

	req.Formato = normalizarFormatoEntrega(req.Formato)
	if req.Formato == "" {
		return nil, errors.New("formato de entrega inválido")
	}

	var resp APIResponse[*ConsultaDocumentoPJEResponse]

	if err := c.http.PostJSON(ctx, routeConsultaDocumento, nil, req, &resp); err != nil {
		return nil, err
	}

	return parseAPIResponse(resp, "erro ao consultar conteúdo de documento")
}

func validarRequisicaoBase(req ConsultaProcessoPJERequest) error {
	if strings.TrimSpace(req.NumeroProcesso) == "" {
		return errors.New("numero_processo não informado")
	}

	if strings.TrimSpace(req.UsuarioCpf) == "" {
		return errors.New("usuario_cpf não informado")
	}

	if strings.TrimSpace(req.UsuarioSenha) == "" {
		return errors.New("usuario_senha não informada")
	}

	return nil
}

func validarRequisicaoDocumento(req ConsultaProcessoPJERequest) error {
	if err := validarRequisicaoBase(req); err != nil {
		return err
	}

	if len(limparStrings(req.Documentos)) == 0 {
		return errors.New("informe ao menos um idDocumento no campo documentos")
	}

	return nil
}

func normalizarFormatoEntrega(formato string) string {
	formato = strings.ToLower(strings.TrimSpace(formato))

	if formato == "" {
		return FormatoEntregaAuto
	}

	switch formato {
	case FormatoEntregaAuto,
		FormatoEntregaHTML,
		FormatoEntregaTexto,
		FormatoEntregaBase64,
		FormatoEntregaMetadados:
		return formato
	default:
		return ""
	}
}

func limparStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if seen[value] {
			continue
		}

		seen[value] = true
		out = append(out, value)
	}

	return out
}

func parseAPIResponse[T any](
	resp APIResponse[*T],
	fallback string,
) (*T, error) {
	if !resp.OK {
		return nil, buildAPIError(resp.Error, resp.Message, fallback)
	}

	if resp.Data == nil {
		return nil, fmt.Errorf("mni-srv retornou resposta sem data")
	}

	return resp.Data, nil
}

func buildAPIError(apiErr *APIError, message string, fallback string) error {
	if apiErr != nil {
		if strings.TrimSpace(apiErr.Description) != "" {
			return errors.New(apiErr.Description)
		}

		if strings.TrimSpace(apiErr.Message) != "" {
			return errors.New(apiErr.Message)
		}
	}

	if strings.TrimSpace(message) != "" {
		return errors.New(message)
	}

	return fmt.Errorf("%s", fallback)
}
