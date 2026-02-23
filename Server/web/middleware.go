package web

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/MHSarmadi/Umbra/Server/logger"
)

type middleware func(http.Handler) http.Handler

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rw *responseRecorder) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

func (rw *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return hj.Hijack()
}

func (rw *responseRecorder) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (rw *responseRecorder) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := rw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (rw *responseRecorder) ReadFrom(src io.Reader) (int64, error) {
	rf, ok := rw.ResponseWriter.(io.ReaderFrom)
	if !ok {
		return io.Copy(rw.ResponseWriter, src)
	}
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rf.ReadFrom(src)
	rw.bytes += int(n)
	return n, err
}

func chainMiddlewares(h http.Handler, mws ...middleware) http.Handler {
	if len(mws) == 0 {
		return h
	}
	wrapped := h
	for i := len(mws) - 1; i >= 0; i-- {
		wrapped = mws[i](wrapped)
	}
	return wrapped
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Errorf("panic recovered method=%s path=%s remote=%s panic=%v", r.Method, r.URL.Path, r.RemoteAddr, rec)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &responseRecorder{ResponseWriter: w}
		start := time.Now()
		next.ServeHTTP(recorder, r)
		duration := time.Since(start)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}
		logger.Debugf(
			"%s %s from [%s] status=%d bytes=%d duration_ms=%d",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			recorder.status,
			recorder.bytes,
			duration.Milliseconds(),
		)
	})
}
