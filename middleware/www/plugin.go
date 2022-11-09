package www

import (
	"net/http"
	"strings"
)

const PluginName = "www"

type Plugin struct{}

func (p *Plugin) Init() error {
	return nil
}

// Name returns user-friendly plugin name
func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if data := strings.Split(r.URL.Host, "."); len(data) == 2 {
			r.URL.Host = "www." + r.URL.Host
			w.Header().Set("Location", r.URL.String())
			w.WriteHeader(http.StatusMovedPermanently)
		}
		next.ServeHTTP(w, r)
	})
}
