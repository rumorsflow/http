package parallel_requests

import (
	"fmt"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"net/http"
)

const (
	RootPluginName = "http"
	PluginName     = "parallel_requests"
)

type Plugin struct {
	cfg *Config
	sem chan struct{}
}

func (p *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("http parallel requests plugin init")

	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	if !cfg.Has(fmt.Sprintf("%s.%s", RootPluginName, PluginName)) {
		return errors.E(op, errors.Disabled)
	}

	if err := cfg.UnmarshalKey(fmt.Sprintf("%s.%s", RootPluginName, PluginName), &p.cfg); err != nil {
		return errors.E(op, err)
	}

	if p.cfg.MaxAllowed == 0 {
		return errors.E(op, errors.Disabled)
	}

	p.sem = make(chan struct{}, p.cfg.MaxAllowed)

	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.acquire()
		defer p.release()

		next.ServeHTTP(w, r)
	})
}

func (p *Plugin) acquire() {
	p.sem <- struct{}{}
}

func (p *Plugin) release() {
	<-p.sem
}
