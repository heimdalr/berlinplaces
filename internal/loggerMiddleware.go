package internal

import (
	"github.com/rs/zerolog"
	"github.com/urfave/negroni"
	"net"
	"net/http"
	"time"
)

// LoggerMiddleware a type to implement our logging middleware.
type LoggerMiddleware struct {
	Handler http.Handler
	Logger  zerolog.Logger
}

// ServeHTTP implements the Handler interface for our logging middleware.
func (lmw LoggerMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t := time.Now()
	url := *req.URL
	rw := negroni.NewResponseWriter(w)
	lmw.Handler.ServeHTTP(rw, req)
	ip, _, _ := net.SplitHostPort(req.RemoteAddr)
	lmw.Logger.Info().
		Str("remote", ip).
		Str("method", req.Method).
		Str("uri", url.RequestURI()).
		Int64("µs", time.Since(t).Microseconds()).
		Int("status", rw.Status()).
		Int("size", rw.Size()).
		Msg("request")
}
