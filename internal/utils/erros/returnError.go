/*
---------------------------------------------------------------------------------------
File: erros.go
Autor: Aldenor
Inspiração: Enterprise Applications with Gin
Data: 03-05-2025
---------------------------------------------------------------------------------------
*/
package erros

import (
	"fmt"
	"log"
	"math/rand"

	"time"
)

// Função imprime no log o time atual, que pode ser o início ou o fim de um processo
func LogTimeMessage(message string) {
	log.Printf(message+": %s", time.Now().Format("2006-01-02 15:04:05"))
}

func CreateError(message string, details ...string) error {
	return fmt.Errorf("%s: %v", message, details)
}

func CreateErrorf(message string, args ...interface{}) error {

	return fmt.Errorf(message, args...)
}

// backoff simples que calcula e retorna um lapso temporal para repetir uma tentativa que falhou
func RetryBackoff(attempt int) time.Duration {
	// 200ms, 400ms, 800ms, máx 2s + jitter
	base := 200 * time.Millisecond
	d := base << (attempt - 1)
	if d > 2*time.Second {
		d = 2 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(100 * time.Millisecond)))
	return d + jitter
}
