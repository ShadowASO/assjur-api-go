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
