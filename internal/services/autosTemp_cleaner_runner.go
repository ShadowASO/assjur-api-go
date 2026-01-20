package services

/*
File: autosTemp_cleaner_runner.go
Data: 20/01/2026
Finalidade: Faz a limpeza do índice autos_temp, a cada hora, deletando os registros com
mais de 24 horas.

*/

import (
	"context"
	"sync/atomic"
	"time"

	"ocrserver/internal/utils/logger"
)

type AutosTempCleaner struct {
	svc       *AutosTempServiceType
	interval  time.Duration
	olderThan time.Duration

	running atomic.Bool // impede sobreposição
}

func NewAutosTempCleaner(svc *AutosTempServiceType) *AutosTempCleaner {
	return &AutosTempCleaner{
		svc:       svc,
		interval:  time.Hour,
		olderThan: 24 * time.Hour,
	}
}

// Start roda em goroutine. Para parar, cancele o ctx.
func (c *AutosTempCleaner) Start(ctx context.Context) {
	if c == nil || c.svc == nil {
		logger.Log.Error("AutosTempCleaner: svc nil (não iniciado)")
		return
	}
	if c.interval <= 0 {
		logger.Log.Error("AutosTempCleaner: interval inválido")
		return
	}
	if c.olderThan <= 0 {
		logger.Log.Error("AutosTempCleaner: olderThan inválido")
		return
	}

	go func() {
		// Roda uma vez ao iniciar (mantenho seu comportamento)
		c.runOnce(ctx)

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Info("AutosTempCleaner: finalizando (ctx cancelado).")
				return
			case <-ticker.C:
				c.runOnce(ctx)
			}
		}
	}()
}

func (c *AutosTempCleaner) runOnce(ctx context.Context) {
	// Evita concorrência: se uma execução anterior ainda estiver rodando, pula.
	if !c.running.CompareAndSwap(false, true) {
		logger.Log.Warning("AutosTempCleaner: execução anterior ainda em andamento; pulando este ciclo.")
		return
	}
	defer c.running.Store(false)

	now := time.Now()
	start := now
	cutoff := now.Add(-c.olderThan).UTC().Format(time.RFC3339)

	logger.Log.Infof(
		"Iniciando cleanup do índice autos_temp (olderThan=%s, cutoff=%s)",
		c.olderThan.String(),
		cutoff,
	)

	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	deleted, err := c.svc.CleanupOlderThan(runCtx, c.olderThan)
	if err != nil {
		logger.Log.Warningf("AutosTempCleaner: execução com erro: %v", err)
		return
	}

	logger.Log.Infof(
		"Finalizado cleanup do índice autos_temp: removidos=%d, duração=%s",
		deleted,
		time.Since(start).Truncate(time.Millisecond),
	)
}
