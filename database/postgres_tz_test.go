package database

import (
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

// Run with: GALAPLATE_TEST_PG_DSN="host=127.0.0.1 user=postgres password=... dbname=postgres sslmode=disable" go test ./database -run PostgresScanLocation
func TestPostgresScanLocation(t *testing.T) {
	dsn := os.Getenv("GALAPLATE_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("GALAPLATE_TEST_PG_DSN not set")
	}

	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}

	scan := func(loc *time.Location) (tz, ts time.Time) {
		cfg, err := pgx.ParseConfig(dsn)
		if err != nil {
			t.Fatal(err)
		}
		db := openPostgres(cfg, loc)
		defer db.Close()
		err = db.QueryRow(`SELECT '2026-10-09 12:00:00+00'::timestamptz, '2026-10-09 12:00:00'::timestamp`).Scan(&tz, &ts)
		if err != nil {
			t.Fatal(err)
		}
		return tz, ts
	}

	tzJKT, tsJKT := scan(jakarta)
	tzUTC, tsUTC := scan(time.UTC)

	// timestamptz: label follows app tz, instant unchanged.
	if _, off := tzJKT.Zone(); off != 7*3600 {
		t.Errorf("timestamptz offset = %d, want +7h", off)
	}
	if !tzJKT.Equal(tzUTC) {
		t.Errorf("timestamptz instant changed: %v vs %v", tzJKT, tzUTC)
	}

	// timestamp WITHOUT time zone: wall clock is interpreted in app tz.
	if _, off := tsJKT.Zone(); off != 7*3600 {
		t.Errorf("timestamp offset = %d, want +7h", off)
	}
	if tsJKT.Format(time.DateTime) != tsUTC.Format(time.DateTime) {
		t.Errorf("timestamp wall clock changed: %v vs %v", tsJKT, tsUTC)
	}
}
