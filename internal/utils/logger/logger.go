// package logger provides a simple leveled logger with optional file rotation.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Level defines log severity.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	OffLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case OffLevel:
		return "off"
	default:
		return "unknown"
	}
}

func parseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return DebugLevel
	case "info", "":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "off", "none":
		return OffLevel
	default:
		return InfoLevel
	}
}

// LoggerType holds writers and leveled loggers.
type LoggerType struct {
	mw          io.Writer
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	closer      io.Closer // if the writer supports Close()
	level       Level
	mu          sync.RWMutex
}

// Global logger and one-time init guard.
var (
	Log            *LoggerType
	onceInitLogger sync.Once
)

// NewLogger creates a simple file-based logger without rotation.
func NewLogger(logFileName string) (*LoggerType, error) {
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo de log: %w", err)
	}

	// date|time|micros + caller short file:line
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	l := &LoggerType{
		mw:          logFile,
		closer:      logFile,
		level:       InfoLevel,
		debugLogger: log.New(logFile, "[DEBUG] ", flags),
		infoLogger:  log.New(logFile, "[INFO]  ", flags),
		warnLogger:  log.New(logFile, "[WARN]  ", flags),
		errorLogger: log.New(logFile, "[ERROR] ", flags),
	}
	return l, nil
}

// InitLoggerGlobal initializes the global logger with rotation and optional stdout mirroring.
// func InitLoggerGlobal(logFilePath string, includeStdout bool) {
// 	onceInitLogger.Do(func() {
// 		rot := &lumberjack.Logger{
// 			Filename:   logFilePath,
// 			MaxSize:    10, // MB
// 			MaxBackups: 5,
// 			MaxAge:     30,   // days
// 			Compress:   true, // gzip old logs
// 		}

// 		var output io.Writer = rot
// 		if includeStdout {
// 			output = io.MultiWriter(os.Stdout, rot)
// 		}

// 		flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

// 		l := &LoggerType{
// 			mw:          output,
// 			level:       parseLevel(os.Getenv("LOG_LEVEL")),
// 			debugLogger: log.New(output, "[DEBUG] ", flags),
// 			infoLogger:  log.New(output, "[INFO]  ", flags),
// 			warnLogger:  log.New(output, "[WARN]  ", flags),
// 			errorLogger: log.New(output, "[ERROR] ", flags),
// 		}

// 		// Keep a closer if available (lumberjack implements Close())
// 		if c, ok := any(rot).(io.Closer); ok {
// 			l.closer = c
// 		}

// 		Log = l
// 		Log.Infof("Logger inicializado (rotação ativa, stdout=%v, nível=%s).", includeStdout, Log.level)
// 	})
// }

func ensureLogDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func InitLoggerGlobal(logFilePath string, includeStdout bool) {
	onceInitLogger.Do(func() {
		// Garante que a pasta existe
		if err := ensureLogDir(logFilePath); err != nil {
			panic(fmt.Errorf("erro ao criar diretório de log: %w", err))
		}

		rot := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    10, // MB
			MaxBackups: 5,
			MaxAge:     30,   // dias
			Compress:   true, // gzip
		}

		var output io.Writer = rot
		if includeStdout {
			output = io.MultiWriter(os.Stdout, rot)
		}

		flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

		l := &LoggerType{
			mw:          output,
			level:       parseLevel(os.Getenv("LOG_LEVEL")),
			debugLogger: log.New(output, "[DEBUG] ", flags),
			infoLogger:  log.New(output, "[INFO]  ", flags),
			warnLogger:  log.New(output, "[WARN]  ", flags),
			errorLogger: log.New(output, "[ERROR] ", flags),
		}

		if c, ok := any(rot).(io.Closer); ok {
			l.closer = c
		}

		Log = l
		Log.Infof("Logger inicializado (rotação ativa, stdout=%v, nível=%s).", includeStdout, Log.level)
	})
}

// SetLevel changes the minimum level at runtime (thread-safe).
func (l *LoggerType) SetLevel(level Level) {
	if l == nil {
		return
	}
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// Level returns the current level.
func (l *LoggerType) Level() Level {
	if l == nil {
		return InfoLevel
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// --- Helpers to check if a message should be logged for current level ---

func (l *LoggerType) enabled(target Level) bool {
	if l == nil {
		return true
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return target >= l.level && l.level != OffLevel
}

// Call depth so file:line shows the caller site, not the method itself.
// 2 is usually correct here (this method -> Output -> caller).
const callDepth = 2

// Debug logs a debug message.
func (l *LoggerType) Debug(message string) {
	if l == nil {
		if parseLevel(os.Getenv("LOG_LEVEL")) <= DebugLevel {
			log.Println("[DEBUG]", message)
		}
		return
	}
	if !l.enabled(DebugLevel) || l.debugLogger == nil {
		return
	}
	l.debugLogger.Output(callDepth, message)
}

func (l *LoggerType) Debugf(format string, args ...interface{}) {
	if l == nil {
		if parseLevel(os.Getenv("LOG_LEVEL")) <= DebugLevel {
			log.Printf("[DEBUG] "+format, args...)
		}
		return
	}
	if !l.enabled(DebugLevel) || l.debugLogger == nil {
		return
	}
	l.debugLogger.Output(callDepth, fmt.Sprintf(format, args...))
}

// Info logs an info message.
func (l *LoggerType) Info(message string) {
	if l == nil || l.infoLogger == nil {
		log.Println("[INFO] ", message)
		return
	}
	if !l.enabled(InfoLevel) {
		return
	}
	l.infoLogger.Output(callDepth, message)
}

func (l *LoggerType) Infof(format string, args ...interface{}) {
	if l == nil || l.infoLogger == nil {
		log.Printf("[INFO] "+format, args...)
		return
	}
	if !l.enabled(InfoLevel) {
		return
	}
	l.infoLogger.Output(callDepth, fmt.Sprintf(format, args...))
}

// Warning logs a warning message.
func (l *LoggerType) Warning(message string) {
	if l == nil || l.warnLogger == nil {
		log.Println("[WARN] ", message)
		return
	}
	if !l.enabled(WarnLevel) {
		return
	}
	l.warnLogger.Output(callDepth, message)
}

func (l *LoggerType) Warningf(format string, args ...interface{}) {
	if l == nil || l.warnLogger == nil {
		log.Printf("[WARN] "+format, args...)
		return
	}
	if !l.enabled(WarnLevel) {
		return
	}
	l.warnLogger.Output(callDepth, fmt.Sprintf(format, args...))
}

// Error logs an error message with optional details.
func (l *LoggerType) Error(message string, details ...string) {
	if l == nil || l.errorLogger == nil {
		if len(details) > 0 {
			log.Println("[ERROR]", message, "| Detalhes:", strings.Join(details, " | "))
		} else {
			log.Println("[ERROR]", message)
		}
		return
	}
	if !l.enabled(ErrorLevel) {
		return
	}
	fullMessage := message
	if len(details) > 0 {
		fullMessage += " | Detalhes: " + strings.Join(details, " | ")
	}
	l.errorLogger.Output(callDepth, fullMessage)
}

// Errorf logs a formatted error message.
func (l *LoggerType) Errorf(format string, args ...interface{}) {
	if l == nil || l.errorLogger == nil {
		log.Printf("[ERROR] "+format, args...)
		return
	}
	if !l.enabled(ErrorLevel) {
		return
	}
	l.errorLogger.Output(callDepth, fmt.Sprintf(format, args...))
}

// ErrorErr logs an error with context (handy helper).
func (l *LoggerType) ErrorErr(err error, context string) {
	if err == nil {
		return
	}
	l.Errorf("%s | erro=%v", context, err)
}

// Close flushes/closes the underlying writer if it supports Close().
func (l *LoggerType) Close() error {
	if l == nil {
		return nil
	}
	if l.closer != nil {
		return l.closer.Close()
	}
	return nil
}

// --- Convenience globals ---

// SetGlobalLevel allows changing global logger level safely.
func SetGlobalLevel(level Level) {
	if Log != nil {
		Log.SetLevel(level)
	}
}

// SetGlobalLevelFromEnv reads LOG_LEVEL and applies it.
func SetGlobalLevelFromEnv() {
	if Log != nil {
		Log.SetLevel(parseLevel(os.Getenv("LOG_LEVEL")))
	}
}

// ExampleInitDev configures a development logger (stdout only, debug level).
func ExampleInitDev() {
	if Log != nil {
		return
	}
	w := os.Stdout
	flags := log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	Log = &LoggerType{
		mw:          w,
		level:       DebugLevel,
		debugLogger: log.New(w, "[DEBUG] ", flags),
		infoLogger:  log.New(w, "[INFO]  ", flags),
		warnLogger:  log.New(w, "[WARN]  ", flags),
		errorLogger: log.New(w, "[ERROR] ", flags),
	}
	Log.Debug("Logger de desenvolvimento inicializado.")
}

// Timestamp returns a RFC3339 timestamp (useful for ad-hoc messages).
func Timestamp() string {
	return time.Now().Format(time.RFC3339Nano)
}
