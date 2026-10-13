// Package timezone resolves and provides the application timezone.
//
// Resolution precedence (first non-empty wins):
//  1. the `app.timezone` config key (config/app.yaml),
//  2. the APP_TIMEZONE environment variable,
//  3. UTC.
//
// The resolved location is cached for the process lifetime, because env.Get
// re-reads the .env file on every miss and time.LoadLocation hits the IANA
// database on disk. Resolution never panics: an invalid value falls back to
// UTC with a single warning.
//
// SQLite note: the SQLite driver stores whatever it is handed and has no
// location-aware decoding, so the convention is UTC in / UTC out — write
// UTC values and treat reads as UTC. Timezone-aware drivers (MySQL via
// loc=, Postgres via the codecs registered in database/connect.go) get the
// app location injected there instead.
package timezone

import (
	"os"
	"sync"
	"time"

	"github.com/galaplate/core/config"
	"github.com/galaplate/core/logger"
)

const (
	// ConfigKey is the config key holding the IANA timezone name.
	ConfigKey = "app.timezone"
	// EnvKey is the environment variable override.
	EnvKey = "APP_TIMEZONE"
)

var (
	resolveOnce sync.Once
	cachedLoc   *time.Location
)

// Location returns the application timezone.
//
// Precedence: app.timezone config key → APP_TIMEZONE env → UTC.
// An invalid IANA name falls back to UTC with a logger.Warn (never panics).
// The result is cached after the first call.
func Location() *time.Location {
	resolveOnce.Do(func() {
		cachedLoc = resolve()
		logger.SetLocation(cachedLoc)
	})

	return cachedLoc
}

// resolve performs the one-time lookup. Env is read via os.Getenv rather
// than env.Get: the config loader already interpolated .env-backed yaml, and
// env.Get would re-read .env (and print an error) on every miss.
func resolve() *time.Location {
	name := config.ConfigString(ConfigKey)
	if name == "" {
		name = os.Getenv(EnvKey)
	}

	// Empty means UTC, silently.
	if name == "" {
		return time.UTC
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		logger.Warn("timezone: invalid timezone, falling back to UTC", map[string]any{
			"timezone": name,
			"error":    err.Error(),
		})

		return time.UTC
	}

	return loc
}

// Now returns the current time in the application timezone.
func Now() time.Time {
	return time.Now().In(Location())
}

// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
	return time.Now().UTC()
}

// FormatRFC3339 formats t as RFC3339 in the application timezone.
func FormatRFC3339(t time.Time) string {
	return t.In(Location()).Format(time.RFC3339)
}

// FormatDate formats t as a date-only string (YYYY-MM-DD) in the
// application timezone.
func FormatDate(t time.Time) string {
	return t.In(Location()).Format("2006-01-02")
}