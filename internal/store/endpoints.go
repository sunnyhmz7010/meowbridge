package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type EndpointInput struct {
	Name          string
	Token         string
	MeowNickname  string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	ParserConfig  string
	Active        bool
}

type EndpointUpdate struct {
	Name          string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	ParserConfig  string
	Active        bool
}

func (s *Store) CreateEndpoint(ctx context.Context, input EndpointInput) (Endpoint, error) {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO endpoints(name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, parser_config, active, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.Token, input.MeowNickname, input.DefaultTitle, input.MsgType, input.HTMLHeight, input.DefaultURL, input.DefaultImgURL, input.ParserConfig, boolToInt(input.Active), now, now)
	if err != nil {
		return Endpoint{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) ListEndpoints(ctx context.Context) ([]Endpoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, parser_config, active, created_at, updated_at
		FROM endpoints ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []Endpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}

func (s *Store) GetEndpoint(ctx context.Context, id int64) (Endpoint, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, parser_config, active, created_at, updated_at
		FROM endpoints WHERE id = ?
	`, id)
	return scanEndpoint(row)
}

func (s *Store) GetEndpointByToken(ctx context.Context, token string) (Endpoint, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, parser_config, active, created_at, updated_at
		FROM endpoints WHERE token = ?
	`, token)
	return scanEndpoint(row)
}

func (s *Store) UpdateEndpoint(ctx context.Context, id int64, input EndpointUpdate) (Endpoint, error) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE endpoints
		SET name = ?, default_title = ?, msg_type = ?, html_height = ?, default_url = ?, default_img_url = ?, parser_config = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, input.Name, input.DefaultTitle, input.MsgType, input.HTMLHeight, input.DefaultURL, input.DefaultImgURL, input.ParserConfig, boolToInt(input.Active), time.Now().UTC(), id)
	if err != nil {
		return Endpoint{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) SetEndpointActive(ctx context.Context, id int64, active bool) error {
	result, err := s.db.ExecContext(ctx, `UPDATE endpoints SET active = ?, updated_at = ? WHERE id = ?`, boolToInt(active), time.Now().UTC(), id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

func (s *Store) ResetEndpointToken(ctx context.Context, id int64, newToken string) (Endpoint, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE endpoints SET token = ?, updated_at = ? WHERE id = ?`, newToken, time.Now().UTC(), id)
	if err != nil {
		return Endpoint{}, err
	}
	if err := ensureRowsAffected(result); err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) DeleteEndpoint(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM endpoints WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return ensureRowsAffected(result)
}

type endpointScanner interface {
	Scan(dest ...any) error
}

func scanEndpoint(row endpointScanner) (Endpoint, error) {
	var ep Endpoint
	var active int
	err := row.Scan(&ep.ID, &ep.Name, &ep.Token, &ep.MeowNickname, &ep.DefaultTitle, &ep.MsgType, &ep.HTMLHeight, &ep.DefaultURL, &ep.DefaultImgURL, &ep.ParserConfig, &active, &ep.CreatedAt, &ep.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Endpoint{}, ErrNotFound
	}
	if err != nil {
		return Endpoint{}, err
	}
	ep.Active = active == 1
	return ep, nil
}

func ensureRowsAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
