package headers

import (
	"fmt"
	"github.com/roadrunner-server/errors"
	"github.com/rumorsflow/contracts/config"
	"net/http"
	"strconv"
)

const (
	RootPluginName = "http"
	PluginName     = "headers"
)

type Plugin struct {
	cfg *Config
}

func (p *Plugin) Init(cfg config.Configurer) error {
	const op = errors.Op("http headers plugin init")

	if !cfg.Has(RootPluginName) {
		return errors.E(op, errors.Disabled)
	}

	if !cfg.Has(fmt.Sprintf("%s.%s", RootPluginName, PluginName)) {
		return errors.E(op, errors.Disabled)
	}

	if err := cfg.UnmarshalKey(fmt.Sprintf("%s.%s", RootPluginName, PluginName), &p.cfg); err != nil {
		return errors.E(op, err)
	}

	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.cfg.Request != nil {
			for k, v := range p.cfg.Request {
				r.Header.Add(k, v)
			}
		}

		if p.cfg.Response != nil {
			for k, v := range p.cfg.Response {
				w.Header().Set(k, v)
			}
		}

		if p.cfg.CORS != nil {
			if r.Method == http.MethodOptions {
				p.preflightRequest(w)

				return
			}
			p.corsHeaders(w)
		}

		next.ServeHTTP(w, r)
	})
}

// configure OPTIONS response
func (p *Plugin) preflightRequest(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")

	if p.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", p.cfg.CORS.AllowedOrigin)
	}

	if p.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", p.cfg.CORS.AllowedHeaders)
	}

	if p.cfg.CORS.AllowedMethods != "" {
		headers.Set("Access-Control-Allow-Methods", p.cfg.CORS.AllowedMethods)
	}

	if p.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*p.cfg.CORS.AllowCredentials))
	}

	if p.cfg.CORS.MaxAge > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(p.cfg.CORS.MaxAge))
	}

	w.WriteHeader(http.StatusOK)
}

// configure CORS headers
func (p *Plugin) corsHeaders(w http.ResponseWriter) {
	headers := w.Header()

	headers.Add("Vary", "Origin")

	if p.cfg.CORS.AllowedOrigin != "" {
		headers.Set("Access-Control-Allow-Origin", p.cfg.CORS.AllowedOrigin)
	}

	if p.cfg.CORS.AllowedHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", p.cfg.CORS.AllowedHeaders)
	}

	if p.cfg.CORS.ExposedHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", p.cfg.CORS.ExposedHeaders)
	}

	if p.cfg.CORS.AllowCredentials != nil {
		headers.Set("Access-Control-Allow-Credentials", strconv.FormatBool(*p.cfg.CORS.AllowCredentials))
	}
}
