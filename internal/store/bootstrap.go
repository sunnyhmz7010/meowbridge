package store

import (
	"context"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type BootstrapOptions struct {
	AdminPassword    string
	MeowAPIBaseURL   string
	LogRetentionDays int
}

func (s *Store) Bootstrap(ctx context.Context, opts BootstrapOptions) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if opts.AdminPassword == "" {
			return errors.New("ADMIN_PASSWORD is required for initial bootstrap")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO admin_users(password_hash, created_at, updated_at)
			VALUES(?, ?, ?)
		`, string(hash), time.Now().UTC(), time.Now().UTC()); err != nil {
			return err
		}
	}
	if err := s.insertSettingIfMissing(ctx, "meow_api_base_url", opts.MeowAPIBaseURL); err != nil {
		return err
	}
	return s.insertSettingIfMissing(ctx, "log_retention_days", strconv.Itoa(opts.LogRetentionDays))
}

func (s *Store) insertSettingIfMissing(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings(key, value, updated_at)
		VALUES(?, ?, ?)
		ON CONFLICT(key) DO NOTHING
	`, key, value, time.Now().UTC())
	return err
}
