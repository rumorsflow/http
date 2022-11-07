package http

import (
	"context"
	"github.com/alexedwards/flow"
	"github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

const PluginName = "http"

type Plugin struct {
	cfg      *Config
	log      *zap.Logger
	mux      *flow.Mux
	srv      *http.Server
	mdwr     *sync.Map
	handlers *sync.Map
	stop     chan struct{}
	once     sync.Once
	grace    time.Duration
}

type Middleware interface {
	Handle(http.Handler) http.Handler
}

type Handler interface {
	Register(mux *flow.Mux)
}

func (p *Plugin) Init(cfg config.Configurer, log *zap.Logger) error {
	const op = errors.Op("http plugin init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	if err := cfg.UnmarshalKey(PluginName, &p.cfg); err != nil {
		return errors.E(op, errors.Init, err)
	}
	p.cfg.InitDefault()

	p.log = log
	p.mux = flow.New()
	p.mdwr = new(sync.Map)
	p.handlers = new(sync.Map)
	p.stop = make(chan struct{}, 1)
	p.grace = cfg.GracefulTimeout()

	return nil
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)

	p.once.Do(func() {
		l := p.log.Sugar()

		for _, key := range p.cfg.Middleware {
			if h, ok := p.mdwr.Load(key); ok {
				l.Infof("register middleware %s", key)
				p.mux.Use(h.(Middleware).Handle)
			}
		}

		p.handlers.Range(func(key, value any) bool {
			l.Infof("register handler %s", key)
			value.(Handler).Register(p.mux)
			return true
		})
	})

	go p.serve(errCh)
	go p.shutdown(errCh)

	return errCh
}

func (p *Plugin) Stop() error {
	p.stop <- struct{}{}
	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) serve(errCh chan error) {
	var err error
	const op = errors.Op("http plugin serve")

	p.srv = p.cfg.Server(p.mux)

	p.log.Info("HTTP server starting")
	if p.cfg.IsTLS() {
		err = p.srv.ListenAndServeTLS(p.cfg.CertFile, p.cfg.KeyFile)
	} else {
		err = p.srv.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		errCh <- errors.E(op, errors.Serve, err)
	}
}

func (p *Plugin) shutdown(errCh chan error) {
	<-p.stop

	const op = errors.Op("http plugin stop")

	p.log.Info("HTTP server stopping")
	defer p.log.Info("HTTP server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), p.grace)
	defer cancel()

	if err := p.srv.Shutdown(ctx); err != nil {
		errCh <- errors.E(op, errors.Stop, err)
	}
}

func (p *Plugin) Collects() []any {
	return []any{
		p.AddMiddleware,
		p.AddHandler,
	}
}

func (p *Plugin) AddMiddleware(name endure.Named, middleware Middleware) {
	p.mdwr.Store(name.Name(), middleware)
}

func (p *Plugin) AddHandler(name endure.Named, handler Handler) {
	p.handlers.Store(name.Name(), handler)
}
