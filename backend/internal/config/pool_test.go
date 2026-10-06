package config

import "testing"

// The pool must stay well under the Postgres default max_connections=100,
// because the admin app shares the same database.
func TestPoolDefaultsLeaveRoomForTheAdminApp(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	Load()

	if AppConfig.DBMaxOpenConns <= 0 {
		t.Fatalf("DBMaxOpenConns = %d, want a positive cap", AppConfig.DBMaxOpenConns)
	}
	if AppConfig.DBMaxOpenConns > 50 {
		t.Errorf("DBMaxOpenConns = %d, too high for a shared max_connections=100", AppConfig.DBMaxOpenConns)
	}
	if AppConfig.DBMaxIdleConns > AppConfig.DBMaxOpenConns {
		t.Errorf("DBMaxIdleConns (%d) is above DBMaxOpenConns (%d)", AppConfig.DBMaxIdleConns, AppConfig.DBMaxOpenConns)
	}
}

func TestPoolSizeIsEnvOverridable(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "7")
	t.Setenv("DB_MAX_IDLE_CONNS", "3")
	Load()

	if AppConfig.DBMaxOpenConns != 7 || AppConfig.DBMaxIdleConns != 3 {
		t.Fatalf("got open=%d idle=%d, want 7 and 3", AppConfig.DBMaxOpenConns, AppConfig.DBMaxIdleConns)
	}
}

func TestPoolSizeIgnoresNonsenseValues(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "not-a-number")
	t.Setenv("DB_MAX_IDLE_CONNS", "0")
	Load()

	if AppConfig.DBMaxOpenConns != 25 || AppConfig.DBMaxIdleConns != 10 {
		t.Fatalf("got open=%d idle=%d, want the defaults 25 and 10", AppConfig.DBMaxOpenConns, AppConfig.DBMaxIdleConns)
	}
}
