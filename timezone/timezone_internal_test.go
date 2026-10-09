package timezone

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/galaplate/core/config"
)

// resetCache resets the package-level sync.Once + cached location so each
// test can exercise a different resolution path.
func resetCache() {
	resolveOnce = sync.Once{}
	cachedLoc = nil
}

func TestLocationDefaultsToUTC(t *testing.T) {
	resetCache()
	os.Unsetenv(EnvKey)
	config.GetGlobal().Set(ConfigKey, "")

	loc := Location()
	if loc != time.UTC {
		t.Fatalf("expected UTC, got %v", loc)
	}
}

func TestLocationFromEnv(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Asia/Jakarta")

	loc := Location()
	_, offset := time.Now().In(loc).Zone()
	if offset != 7*60*60 {
		t.Fatalf("expected +07:00 offset, got %d", offset)
	}
}

func TestLocationConfigBeatsEnv(t *testing.T) {
	resetCache()
	t.Setenv(EnvKey, "Asia/Jakarta")
	config.GetGlobal().Set(ConfigKey, "Asia/Makassar")
	defer config.GetGlobal().Set(ConfigKey, "")

	loc := Location()
	_, offset := time.Now().In(loc).Zone()
	if offset != 8*60*60 {
		t.Fatalf("config app.timezone should win over env; offset = %d", offset)
	}
}

func TestLocationInvalidFallsBack(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Not/A/Real_Zone")

	loc := Location()
	if loc != time.UTC {
		t.Fatalf("invalid value should fall back to UTC, got %v", loc)
	}
}

func TestNowRespectsLocation(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Asia/Jakarta")

	now := Now()
	if now.Location() != Location() {
		t.Fatalf("Now() location mismatch: %v", now.Location())
	}
	_, offset := now.Zone()
	if offset != 7*60*60 {
		t.Fatalf("expected +07:00, got %d", offset)
	}
}

func TestNowUTC(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Asia/Jakarta")

	if loc := NowUTC().Location(); loc != time.UTC {
		t.Fatalf("NowUTC should be UTC, got %v", loc)
	}
}

func TestFormatRFC3339(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Asia/Jakarta")

	instant := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	if got, want := FormatRFC3339(instant), "2026-10-09T19:00:00+07:00"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestFormatDate(t *testing.T) {
	resetCache()
	config.GetGlobal().Set(ConfigKey, "")
	t.Setenv(EnvKey, "Asia/Jakarta")

	edge := time.Date(2026, 10, 9, 17, 30, 0, 0, time.UTC)
	if got, want := FormatDate(edge), "2026-10-10"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
