package utils

import (
	"context"
	"errors"
	"net"

	"strings"
)

// IsClientAbortErr verifica se o erro foi causado pelo cliente encerrar a requisição
func IsClientAbortErr(err error) bool {
	if err == nil {
		return false
	}

	// Cancelamento explícito via contexto
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Erros típicos de conexão encerrada
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Timeout ou conexão resetada
		if netErr.Timeout() || strings.Contains(strings.ToLower(err.Error()), "use of closed network connection") {
			return true
		}
	}

	// Erros HTTP do servidor que indicam cancelamento
	if strings.Contains(strings.ToLower(err.Error()), "client disconnected") ||
		strings.Contains(strings.ToLower(err.Error()), "broken pipe") ||
		strings.Contains(strings.ToLower(err.Error()), "connection reset by peer") {
		return true
	}

	return false
}
