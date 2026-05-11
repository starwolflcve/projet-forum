package security

import "time"

type AppConfig struct {
	ServerPort        string
	DatabasePath      string
	UploadDir         string
	MaxUploadSize     int64
	SessionCookieName string
	TLSCertFile       string
	TLSKeyFile        string
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

func LoadConfig() AppConfig {
	return AppConfig{
		ServerPort:        ":443",
		DatabasePath:      "./forum.db",
		UploadDir:         "./web/static/uploads",
		MaxUploadSize:     20 * 1024 * 1024,
		TLSCertFile:       "./certs/cert.pem",
		TLSKeyFile:        "./certs/key.pem",
		RateLimitRequests: 50,
		RateLimitWindow:   time.Minute,
	}
}
