package recovery

import (
	"fmt"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	RootPluginName = "http"
	PluginName     = "recovery"
)

type ErrorHandler interface {
	Handle(w http.ResponseWriter, err error)
}

type Plugin struct {
	log        *zap.Logger
	errHandler ErrorHandler
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger, errHandler ErrorHandler) error {
	const op = errors.Op("http recovery plugin init")

	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	p.log = log
	p.errHandler = errHandler

	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(r, false)
				if brokenPipe {
					p.log.Error(r.URL.Path,
						zap.Any("error", err),
						zap.ByteString("request", httpRequest),
					)
					return
				}

				stack := make([]byte, 4<<10) // 4KB
				length := runtime.Stack(stack, true)
				stack = stack[:length]

				if _, ok := err.(error); !ok {
					err = fmt.Errorf("%v", err)
				}

				p.log.Error("recovery from panic",
					zap.Time("time", time.Now()),
					zap.Error(err.(error)),
					zap.ByteString("request", httpRequest),
					zap.ByteString("stack", stack),
				)
				p.errHandler.Handle(w, err.(error))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
