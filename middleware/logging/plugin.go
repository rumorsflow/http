package logging

import (
	"context"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"time"
)

const (
	RootPluginName = "http"
	PluginName     = "logging"
)

type Plugin struct {
	log *zap.Logger
}

type ctxKey struct{}

func CtxLog(ctx context.Context) *zap.Logger {
	if v, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok {
		return v
	}
	return nil
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("http logging plugin init")

	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	p.log = log

	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		method := r.Method
		host := r.URL.Host
		path := r.URL.Path
		query := r.URL.RawQuery
		uri := r.URL.String()

		if host == "" {
			host = r.Host
		}

		ww := &wrapperResponseWriter{ResponseWriter: w}

		next.ServeHTTP(ww, r.WithContext(context.WithValue(r.Context(), ctxKey{}, p.log)))

		end := time.Now()
		latency := end.Sub(start)

		fields := []zapcore.Field{
			zap.Int("status", ww.statusCode),
			zap.String("method", method),
			zap.String("host", host),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", r.RemoteAddr),
			zap.String("user-agent", r.UserAgent()),
			zap.Duration("latency", latency),
			zap.String("time", end.UTC().Format(time.RFC3339)),
		}

		p.log.Info(uri, fields...)
	})
}

type wrapperResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrapperResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
