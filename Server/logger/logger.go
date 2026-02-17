package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogVerbosity controls which logs are emitted.
// 0 = production silent mode (no logs).
// Higher values enable more logs.
const LogVerbosity uint8 = 3

const (
	LevelError uint8 = 1
	LevelInfo  uint8 = 2
	LevelDebug uint8 = 3
)

var (
	writerMu sync.Mutex
	writer   *dateFileWriter
)

type dateFileWriter struct {
	mu      sync.Mutex
	logDir  string
	curDay  string
	curFile *os.File
}

func newDateFileWriter(logDir string) (*dateFileWriter, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	return &dateFileWriter{logDir: logDir}, nil
}

func (w *dateFileWriter) ensureFileLocked() error {
	day := time.Now().UTC().Format("2006-01-02")
	if w.curFile != nil && w.curDay == day {
		return nil
	}

	if w.curFile != nil {
		if err := w.curFile.Close(); err != nil {
			return err
		}
		w.curFile = nil
	}

	path := filepath.Join(w.logDir, day+".log")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.curFile = f
	w.curDay = day
	return nil
}

func (w *dateFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.ensureFileLocked(); err != nil {
		return 0, err
	}
	return w.curFile.Write(p)
}

func (w *dateFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.curFile == nil {
		return nil
	}
	err := w.curFile.Close()
	w.curFile = nil
	w.curDay = ""
	return err
}

func Init(logDir string) error {
	writerMu.Lock()
	defer writerMu.Unlock()

	if LogVerbosity == 0 {
		log.SetOutput(io.Discard)
		return nil
	}

	w, err := newDateFileWriter(logDir)
	if err != nil {
		return err
	}

	writer = w
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.LUTC)
	log.SetOutput(io.MultiWriter(os.Stdout, writer))
	return nil
}

func Close() error {
	writerMu.Lock()
	defer writerMu.Unlock()

	if writer == nil {
		return nil
	}
	err := writer.Close()
	writer = nil
	return err
}

func enabled(level uint8) bool {
	if LogVerbosity == 0 {
		return false
	}
	return level <= LogVerbosity
}

func logf(level uint8, label, format string, args ...any) {
	if !enabled(level) {
		return
	}
	log.Printf("[%s] %s", label, fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	logf(LevelError, "ERROR", format, args...)
}

func Infof(format string, args ...any) {
	logf(LevelInfo, "INFO", format, args...)
}

func Debugf(format string, args ...any) {
	logf(LevelDebug, "DEBUG", format, args...)
}
