/*
---------------------------------------------------------------------------------------
File: logger.go
Autor: Aldenor
Inspiração: Mastering Go e GPT4o
Data: 05-05-2025
---------------------------------------------------------------------------------------
Este módulo utiliza o pacakage 'lumberjack' para criar e gerenciar a rotação do arquivo
de log quando ele atinge um determinado tamanho ou tempo de criação.
O que acontece automaticamente:

logs/app.log será criado.
Quando atingir 10 MB, será rotacionado: app.log.1.gz, app.log.2.gz, ...
Apenas os 5 logs mais recentes serão mantidos (por até 30 dias).
Logs antigos serão compactados com gzip.

*/

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerType representa a configuração de log da aplicação
type LoggerType struct {
	logFile     *os.File
	mw          io.Writer
	mutex       sync.Mutex
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

var Log *LoggerType
var onceInitLogger sync.Once

// NewLogger cria uma instância de logger, mas não a define como global
func NewLogger(logFileName string) (*LoggerType, error) {
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo de log: %w", err)
	}

	multiWriter := io.MultiWriter(logFile)
	return &LoggerType{
		logFile:     logFile,
		mw:          multiWriter,
		infoLogger:  log.New(multiWriter, "[INFO] ", log.LstdFlags|log.Lshortfile),
		warnLogger:  log.New(multiWriter, "[WARNING] ", log.LstdFlags|log.Lshortfile),
		errorLogger: log.New(multiWriter, "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}, nil
}

// InitGlobalLogger inicializa o logger padrão global com fallback para stdout
func InitGlobalLogger(logFilePath string, includeStdout bool) {
	onceInitLogger.Do(func() {
		rotatingWriter := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     30,   // dias
			Compress:   true, // compacta os antigos
		}

		var output io.Writer
		if includeStdout {
			output = io.MultiWriter(os.Stdout, rotatingWriter)
		} else {
			output = rotatingWriter
		}

		Log = &LoggerType{
			mw:          output,
			infoLogger:  log.New(output, "[INFO] ", log.LstdFlags|log.Lshortfile),
			warnLogger:  log.New(output, "[WARNING] ", log.LstdFlags|log.Lshortfile),
			errorLogger: log.New(output, "[ERROR] ", log.LstdFlags|log.Lshortfile),
		}

		Log.Info("Logger com rotação configurado com sucesso.")
	})
}

// Info registra uma mensagem informativa
func (l *LoggerType) Info(message string) {
	if l == nil || l.infoLogger == nil {
		log.Println("[INFO]", message)
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.infoLogger.Output(2, message)
}

// Warning registra uma mensagem de aviso
func (l *LoggerType) Warning(message string) {
	if l == nil || l.warnLogger == nil {
		log.Println("[WARNING]", message)
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.warnLogger.Output(2, message)
}

// Error registra uma mensagem de erro com detalhes opcionais
func (l *LoggerType) Error(message string, details ...string) {
	if l == nil || l.errorLogger == nil {
		log.Println("[ERROR]", message, strings.Join(details, " | "))
		return
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	fullMessage := message
	if len(details) > 0 {
		fullMessage += " | Detalhes: " + strings.Join(details, " | ")
	}
	l.errorLogger.Output(2, fullMessage)
}

// Close fecha o arquivo de log de forma segura
func (l *LoggerType) Close() {
	//SEm uso com o lumberjack
}
