package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PushLogInput struct {
	EndpointID       int64
	EndpointName     string
	Token            string
	SourceType       string
	RequestMethod    string
	RequestHeaders   string
	RequestQuery     string
	RequestPayload   string
	ParsedTitle      string
	ParsedMsg        string
	ParsedMsgType    string
	MeowStatusCode   int
	MeowResponseBody string
	Success          bool
	ErrorMessage     string
}

func (s *Store) CreatePushLog(ctx context.Context, input PushLogInput) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO push_logs(endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.EndpointID, input.EndpointName, input.Token, input.SourceType, input.RequestMethod, input.RequestHeaders, input.RequestQuery, input.RequestPayload, input.ParsedTitle, input.ParsedMsg, input.ParsedMsgType, input.MeowStatusCode, input.MeowResponseBody, boolToInt(input.Success), input.ErrorMessage, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetPushLog(ctx context.Context, id int64) (PushLog, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at
		FROM push_logs WHERE id = ?
	`, id)
	log, err := scanPushLog(row)
	if errors.Is(err, sql.ErrNoRows) {
		return PushLog{}, ErrNotFound
	}
	return log, err
}

func (s *Store) ListPushLogs(ctx context.Context) ([]PushLog, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at
		FROM push_logs ORDER BY id DESC LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []PushLog
	for rows.Next() {
		log, err := scanPushLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (s *Store) CleanupPushLogs(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM push_logs WHERE created_at < ?`, before.UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type pushLogScanner interface {
	Scan(dest ...any) error
}

func scanPushLog(row pushLogScanner) (PushLog, error) {
	var log PushLog
	var success int
	err := row.Scan(&log.ID, &log.EndpointID, &log.EndpointName, &log.Token, &log.SourceType, &log.RequestMethod, &log.RequestHeaders, &log.RequestQuery, &log.RequestPayload, &log.ParsedTitle, &log.ParsedMsg, &log.ParsedMsgType, &log.MeowStatusCode, &log.MeowResponseBody, &success, &log.ErrorMessage, &log.CreatedAt)
	if err != nil {
		return PushLog{}, err
	}
	log.Success = success == 1
	return log, nil
}
