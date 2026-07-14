package store

import (
	"context"
	"testing"
)

func TestMigrateCreatesCoreTables(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	for _, table := range []string{"admin_users", "endpoints", "settings", "push_logs"} {
		var name string
		err := st.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s was not created: %v", table, err)
		}
	}
}

func TestBootstrapCreatesAdminAndSettingsOnce(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	err := st.Bootstrap(ctx, BootstrapOptions{
		AdminPassword:    "first-password",
		MeowAPIBaseURL:   "https://meow.example.test",
		LogRetentionDays: 14,
	})
	if err != nil {
		t.Fatalf("Bootstrap first: %v", err)
	}

	err = st.Bootstrap(ctx, BootstrapOptions{
		AdminPassword:    "second-password",
		MeowAPIBaseURL:   "https://changed.example.test",
		LogRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("Bootstrap second: %v", err)
	}

	var adminCount int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&adminCount); err != nil {
		t.Fatalf("count admins: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("adminCount = %d", adminCount)
	}

	baseURL, err := st.GetSetting(ctx, "meow_api_base_url")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if baseURL != "https://meow.example.test" {
		t.Fatalf("baseURL = %q", baseURL)
	}
}

func TestBootstrapRequiresAdminPasswordForInitialAdmin(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	err := st.Bootstrap(ctx, BootstrapOptions{
		MeowAPIBaseURL:   "https://meow.example.test",
		LogRetentionDays: 14,
	})
	if err == nil {
		t.Fatal("Bootstrap() error = nil")
	}

	var adminCount int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&adminCount); err != nil {
		t.Fatalf("count admins: %v", err)
	}
	if adminCount != 0 {
		t.Fatalf("adminCount = %d", adminCount)
	}
}

func openTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	ctx := context.Background()
	st, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return st, func() { _ = st.Close() }
}
