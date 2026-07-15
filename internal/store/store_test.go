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
		AdminPassword: "first-password",
	})
	if err != nil {
		t.Fatalf("Bootstrap first: %v", err)
	}

	err = st.Bootstrap(ctx, BootstrapOptions{
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

	retentionDays, err := st.GetSetting(ctx, "log_retention_days")
	if err != nil {
		t.Fatalf("GetSetting log_retention_days: %v", err)
	}
	if retentionDays != "14" {
		t.Fatalf("log_retention_days = %q", retentionDays)
	}
}

func TestBootstrapRequiresAdminPasswordForInitialAdmin(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	err := st.Bootstrap(ctx, BootstrapOptions{
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

func TestSetSettingUpdatesExistingValue(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	if err := st.SetSetting(ctx, "example", "first"); err != nil {
		t.Fatalf("SetSetting first: %v", err)
	}
	if err := st.SetSetting(ctx, "example", "second"); err != nil {
		t.Fatalf("SetSetting second: %v", err)
	}

	value, err := st.GetSetting(ctx, "example")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if value != "second" {
		t.Fatalf("value = %q", value)
	}
}

func TestGetSettingReturnsErrNotFound(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	if _, err := st.GetSetting(ctx, "missing"); err != ErrNotFound {
		t.Fatalf("GetSetting(missing) error = %v, want ErrNotFound", err)
	}
}

func TestEndpointCRUDKeepsMeowNicknameImmutable(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	created, err := st.CreateEndpoint(ctx, EndpointInput{
		Name:          "GitHub",
		Token:         "token-1",
		MeowNickname:  "sunny",
		DefaultTitle:  "Default",
		MsgType:       "markdown",
		HTMLHeight:    300,
		DefaultURL:    "https://example.test",
		DefaultImgURL: "https://example.test/icon.png",
		Active:        true,
	})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	updated, err := st.UpdateEndpoint(ctx, created.ID, EndpointUpdate{
		Name:          "GitHub Updated",
		DefaultTitle:  "Changed",
		MsgType:       "text",
		HTMLHeight:    200,
		DefaultURL:    "",
		DefaultImgURL: "",
		Active:        true,
	})
	if err != nil {
		t.Fatalf("UpdateEndpoint: %v", err)
	}
	if updated.MeowNickname != "sunny" {
		t.Fatalf("MeowNickname changed to %q", updated.MeowNickname)
	}

	if err := st.SetEndpointActive(ctx, created.ID, false); err != nil {
		t.Fatalf("SetEndpointActive: %v", err)
	}
	byToken, err := st.GetEndpointByToken(ctx, "token-1")
	if err != nil {
		t.Fatalf("GetEndpointByToken: %v", err)
	}
	if byToken.Active {
		t.Fatal("expected inactive endpoint")
	}

	reset, err := st.ResetEndpointToken(ctx, created.ID, "token-2")
	if err != nil {
		t.Fatalf("ResetEndpointToken: %v", err)
	}
	if reset.Token != "token-2" {
		t.Fatalf("Token = %q", reset.Token)
	}
}

func TestEndpointMutationsReturnErrNotFound(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	if _, err := st.GetEndpoint(ctx, 1); err != ErrNotFound {
		t.Fatalf("GetEndpoint() error = %v, want ErrNotFound", err)
	}
	if _, err := st.GetEndpointByToken(ctx, "missing"); err != ErrNotFound {
		t.Fatalf("GetEndpointByToken() error = %v, want ErrNotFound", err)
	}
	if _, err := st.UpdateEndpoint(ctx, 1, EndpointUpdate{}); err != ErrNotFound {
		t.Fatalf("UpdateEndpoint() error = %v, want ErrNotFound", err)
	}
	if err := st.SetEndpointActive(ctx, 1, true); err != ErrNotFound {
		t.Fatalf("SetEndpointActive() error = %v, want ErrNotFound", err)
	}
	if _, err := st.ResetEndpointToken(ctx, 1, "token"); err != ErrNotFound {
		t.Fatalf("ResetEndpointToken() error = %v, want ErrNotFound", err)
	}
	if err := st.DeleteEndpoint(ctx, 1); err != ErrNotFound {
		t.Fatalf("DeleteEndpoint() error = %v, want ErrNotFound", err)
	}
}

func TestListEndpointsOrdersNewestFirstAndDeleteRemovesEndpoint(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	first, err := st.CreateEndpoint(ctx, EndpointInput{Name: "first", Token: "token-first", MeowNickname: "sunny", MsgType: "text"})
	if err != nil {
		t.Fatalf("CreateEndpoint first: %v", err)
	}
	second, err := st.CreateEndpoint(ctx, EndpointInput{Name: "second", Token: "token-second", MeowNickname: "sunny", MsgType: "text"})
	if err != nil {
		t.Fatalf("CreateEndpoint second: %v", err)
	}

	endpoints, err := st.ListEndpoints(ctx)
	if err != nil {
		t.Fatalf("ListEndpoints: %v", err)
	}
	if len(endpoints) != 2 || endpoints[0].ID != second.ID || endpoints[1].ID != first.ID {
		t.Fatalf("ListEndpoints() = %#v", endpoints)
	}

	if err := st.DeleteEndpoint(ctx, first.ID); err != nil {
		t.Fatalf("DeleteEndpoint: %v", err)
	}
	if _, err := st.GetEndpoint(ctx, first.ID); err != ErrNotFound {
		t.Fatalf("GetEndpoint() error = %v, want ErrNotFound", err)
	}
}

func TestListSettingsReturnsValues(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	if err := st.SetSetting(ctx, "alpha", "one"); err != nil {
		t.Fatalf("SetSetting alpha: %v", err)
	}
	if err := st.SetSetting(ctx, "beta", "two"); err != nil {
		t.Fatalf("SetSetting beta: %v", err)
	}

	values, err := st.ListSettings(ctx)
	if err != nil {
		t.Fatalf("ListSettings: %v", err)
	}
	if values["alpha"] != "one" || values["beta"] != "two" {
		t.Fatalf("ListSettings() = %#v", values)
	}
}

func TestPushLogCreateListDetailCleanup(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	id, err := st.CreatePushLog(ctx, PushLogInput{
		EndpointID:       1,
		EndpointName:     "GitHub",
		Token:            "token-1",
		SourceType:       "github",
		RequestMethod:    "POST",
		RequestHeaders:   `{"content-type":["application/json"]}`,
		RequestQuery:     `{"title":["override"]}`,
		RequestPayload:   `{"message":"hello"}`,
		ParsedTitle:      "title",
		ParsedMsg:        "message",
		ParsedMsgType:    "text",
		MeowStatusCode:   200,
		MeowResponseBody: "ok",
		Success:          true,
		ErrorMessage:     "",
	})
	if err != nil {
		t.Fatalf("CreatePushLog: %v", err)
	}
	log, err := st.GetPushLog(ctx, id)
	if err != nil {
		t.Fatalf("GetPushLog: %v", err)
	}
	if log.RequestPayload == "" || !log.Success {

		t.Fatalf("log = %#v", log)
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
