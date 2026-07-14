package store

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/auth"
)

type BootstrapOptions struct {
	AdminPassword    string
	MeowAPIBaseURL   string
	LogRetentionDays int
}

func (s *Store) Bootstrap(ctx context.Context, opts BootstrapOptions) error {
	_, err := s.GetSetting(ctx, "meow_api_base_url")
	if errors.Is(err, ErrNotFound) && opts.MeowAPIBaseURL == "" {
		return errors.New("MEOW_API_BASE_URL is required for initial bootstrap")
	}
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}

	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return err
	}
	needsAdminPassword := count == 0
	if needsAdminPassword && opts.AdminPassword == "" {
		return errors.New("ADMIN_PASSWORD is required for initial bootstrap")
	}

	if needsAdminPassword {
		hash, err := auth.HashPassword(opts.AdminPassword)
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

func (s *Store) AdminPasswordHash(ctx context.Context) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT password_hash FROM admin_users ORDER BY id ASC LIMIT 1`).Scan(&hash)
	return hash, err
}

func (s *Store) UpdateAdminPasswordHash(ctx context.Context, hash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE admin_users SET password_hash = ?, updated_at = ? WHERE id = (SELECT id FROM admin_users ORDER BY id ASC LIMIT 1)`, hash, time.Now().UTC())
	return err
}

func (s *Store) insertSettingIfMissing(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings(key, value, updated_at)
		VALUES(?, ?, ?)
		ON CONFLICT(key) DO NOTHING
	`, key, value, time.Now().UTC())
	return err
}
