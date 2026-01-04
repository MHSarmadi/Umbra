package web

// import (
// 	"net/http"
// 	"time"

// 	"github.com/rs/zerolog/log"
// )

// func RequestLogger(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		start := time.Now()
// 		next.ServeHTTP(w, r)
// 		log.Info().
// 			Str("method", r.Method).
// 			Str("path", r.URL.Path).
// 			Dur("dur", time.Since(start)).
// 			Msg("")
// 	})
// }
