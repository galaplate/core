package fiber

import (
	corsmw "github.com/gofiber/fiber/v3/middleware/cors"
)

// CORSConfig is the configuration for the CORS middleware.
type CORSConfig = corsmw.Config

// NewCORS creates a new CORS middleware handler.
func NewCORS(config ...CORSConfig) Handler {
	return corsmw.New(config...)
}
