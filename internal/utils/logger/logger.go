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
	mw          io.Writer
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	closer      io.Closer // opcional, para fechar arquivos quando necessário
}

var Log *LoggerType
var onceInitLogger sync.Once

// NewLogger cria uma instância de logger com arquivo simples (sem rotação automática)
func NewLogger(logFileName string) (*LoggerType, error) {
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo de log: %w", err)
	}

	return &LoggerType{
		mw:          logFile,
		closer:      logFile,
		infoLogger:  log.New(logFile, "[INFO] ", log.LstdFlags|log.Lshortfile),
		warnLogger:  log.New(logFile, "[WARNING] ", log.LstdFlags|log.Lshortfile),
		errorLogger: log.New(logFile, "[ERROR] ", log.LstdFlags|log.Lshortfile),
	}, nil
}

// InitLoggerGlobal inicializa o logger global com rotação e opcional saída no stdout
func InitLoggerGlobal(logFilePath string, includeStdout bool) {
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
	l.infoLogger.Output(2, message)
}

// Warning registra uma mensagem de aviso
func (l *LoggerType) Warning(message string) {
	if l == nil || l.warnLogger == nil {
		log.Println("[WARNING]", message)
		return
	}
	l.warnLogger.Output(2, message)
}

// Error registra uma mensagem de erro com detalhes opcionais
func (l *LoggerType) Error(message string, details ...string) {
	if l == nil || l.errorLogger == nil {
		log.Println("[ERROR]", message, strings.Join(details, " | "))
		return
	}
	fullMessage := message
	if len(details) > 0 {
		fullMessage += " | Detalhes: " + strings.Join(details, " | ")
	}
	l.errorLogger.Output(2, fullMessage)
}

// Errorf registra uma mensagem de erro formatada
func (l *LoggerType) Errorf(format string, args ...interface{}) {
	if l == nil || l.errorLogger == nil {
		log.Printf("[ERROR] "+format, args...)
		return
	}
	fullMessage := fmt.Sprintf(format, args...)
	l.errorLogger.Output(2, fullMessage)
}

// Close fecha o recurso do logger quando aplicável
func (l *LoggerType) Close() error {
	if l == nil {
		return nil
	}
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}
