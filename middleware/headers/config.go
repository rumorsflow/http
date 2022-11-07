package headers

// Config declares headers service configuration.
type Config struct {
	// CORS settings.
	CORS *CORSConfig `mapstructure:"cors"`

	// Request headers to add to every request.
	Request map[string]string `mapstructure:"request"`

	// Response headers to add to every response.
	Response map[string]string `mapstructure:"response"`
}

type CORSConfig struct {
	// AllowedOrigin: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AllowedOrigin string `mapstructure:"allowed_origin"`

	// AllowedHeaders: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowedHeaders string `mapstructure:"allowed_headers"`

	// AllowedMethods: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowedMethods string `mapstructure:"allowed_methods"`

	// AllowCredentials https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials *bool `mapstructure:"allow_credentials"`

	// ExposeHeaders:  https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers
	ExposedHeaders string `mapstructure:"exposed_headers"`

	// MaxAge of CORS headers in seconds/
	MaxAge int `mapstructure:"max_age"`
}
