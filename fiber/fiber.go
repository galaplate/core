package fiber

import (
	"github.com/gofiber/fiber/v3"
)

// Types
type Ctx = fiber.Ctx
type App = fiber.App
type Handler = fiber.Handler
type Map = fiber.Map
type Config = fiber.Config
type Error = fiber.Error
type TestConfig = fiber.TestConfig
type ListenConfig = fiber.ListenConfig

// App creation
var New = fiber.New

// HTTP status codes
const (
	StatusContinue           = fiber.StatusContinue
	StatusSwitchingProtocols = fiber.StatusSwitchingProtocols
	StatusProcessing         = fiber.StatusProcessing
	StatusEarlyHints         = fiber.StatusEarlyHints

	StatusOK                   = fiber.StatusOK
	StatusCreated              = fiber.StatusCreated
	StatusAccepted             = fiber.StatusAccepted
	StatusNonAuthoritativeInformation = fiber.StatusNonAuthoritativeInformation
	StatusNoContent            = fiber.StatusNoContent
	StatusResetContent         = fiber.StatusResetContent
	StatusPartialContent       = fiber.StatusPartialContent
	StatusMultiStatus          = fiber.StatusMultiStatus
	StatusAlreadyReported      = fiber.StatusAlreadyReported
	StatusIMUsed               = fiber.StatusIMUsed

	StatusMultipleChoices   = fiber.StatusMultipleChoices
	StatusMovedPermanently  = fiber.StatusMovedPermanently
	StatusFound             = fiber.StatusFound
	StatusSeeOther          = fiber.StatusSeeOther
	StatusNotModified       = fiber.StatusNotModified
	StatusUseProxy          = fiber.StatusUseProxy
	StatusTemporaryRedirect = fiber.StatusTemporaryRedirect
	StatusPermanentRedirect = fiber.StatusPermanentRedirect

	StatusBadRequest                   = fiber.StatusBadRequest
	StatusUnauthorized                 = fiber.StatusUnauthorized
	StatusPaymentRequired              = fiber.StatusPaymentRequired
	StatusForbidden                    = fiber.StatusForbidden
	StatusNotFound                     = fiber.StatusNotFound
	StatusMethodNotAllowed             = fiber.StatusMethodNotAllowed
	StatusNotAcceptable                = fiber.StatusNotAcceptable
	StatusProxyAuthRequired            = fiber.StatusProxyAuthRequired
	StatusRequestTimeout               = fiber.StatusRequestTimeout
	StatusConflict                     = fiber.StatusConflict
	StatusGone                         = fiber.StatusGone
	StatusLengthRequired               = fiber.StatusLengthRequired
	StatusPreconditionFailed           = fiber.StatusPreconditionFailed
	StatusRequestEntityTooLarge        = fiber.StatusRequestEntityTooLarge
	StatusRequestURITooLong            = fiber.StatusRequestURITooLong
	StatusUnsupportedMediaType         = fiber.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable = fiber.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            = fiber.StatusExpectationFailed
	StatusTeapot                       = fiber.StatusTeapot
	StatusMisdirectedRequest           = fiber.StatusMisdirectedRequest
	StatusUnprocessableEntity          = fiber.StatusUnprocessableEntity
	StatusLocked                       = fiber.StatusLocked
	StatusFailedDependency             = fiber.StatusFailedDependency
	StatusTooEarly                     = fiber.StatusTooEarly
	StatusUpgradeRequired              = fiber.StatusUpgradeRequired
	StatusPreconditionRequired         = fiber.StatusPreconditionRequired
	StatusTooManyRequests              = fiber.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  = fiber.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   = fiber.StatusUnavailableForLegalReasons

	StatusInternalServerError           = fiber.StatusInternalServerError
	StatusNotImplemented                = fiber.StatusNotImplemented
	StatusBadGateway                    = fiber.StatusBadGateway
	StatusServiceUnavailable            = fiber.StatusServiceUnavailable
	StatusGatewayTimeout                = fiber.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       = fiber.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         = fiber.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           = fiber.StatusInsufficientStorage
	StatusLoopDetected                  = fiber.StatusLoopDetected
	StatusNotExtended                   = fiber.StatusNotExtended
	StatusNetworkAuthenticationRequired = fiber.StatusNetworkAuthenticationRequired
)
