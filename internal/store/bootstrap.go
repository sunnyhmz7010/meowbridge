package store

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/auth"
)

type BootstrapOptions struct {
	AdminPassword    string
	LogRetentionDays int
}

var ErrAdminAlreadyInitialized = errors.New("admin already initialized")
var ErrBlankAdminPassword = errors.New("admin password is required")

func (s *Store) Bootstrap(ctx context.Context, opts BootstrapOptions) error {
	adminExists, err := s.AdminExists(ctx)
	if err != nil {
		return err
	}

	if !adminExists && opts.AdminPassword != "" {
		if err := s.CreateInitialAdmin(ctx, opts.AdminPassword); err != nil {
			return err
		}
	}
	retentionDays := opts.LogRetentionDays
	if retentionDays <= 0 {
		retentionDays = 14
	}
	return s.insertSettingIfMissing(ctx, "log_retention_days", strconv.Itoa(retentionDays))
}

func (s *Store) AdminExists(ctx context.Context) (bool, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) CreateInitialAdmin(ctx context.Context, password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrBlankAdminPassword
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrAdminAlreadyInitialized
	}
	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO admin_users(password_hash, created_at, updated_at)
		VALUES(?, ?, ?)
	`, string(hash), now, now); err != nil {
		return err
	}
	return tx.Commit()
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
