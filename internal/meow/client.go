package meow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 16 * 1024

type Client struct {
	baseURLProvider func(context.Context) (string, error)
	httpClient      *http.Client
}

type PushRequest struct {
	Nickname   string
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
	HTMLHeight int
}

type PushResponse struct {
	StatusCode int
	Body       string
}

func New(baseURL string, timeout time.Duration) *Client {
	return NewWithBaseURLProvider(func(context.Context) (string, error) {
		return baseURL, nil
	}, timeout)
}

func NewWithBaseURLProvider(provider func(context.Context) (string, error), timeout time.Duration) *Client {
	return &Client{
		baseURLProvider: provider,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Push(ctx context.Context, req PushRequest) (PushResponse, error) {
	baseURL, err := c.baseURLProvider(ctx)
	if err != nil {
		return PushResponse{}, err
	}
	baseURL = strings.TrimSpace(baseURL)
	baseURL = strings.Trim(baseURL, "\"")
	target, err := url.Parse(strings.TrimRight(baseURL, "/") + "/" + url.PathEscape(req.Nickname))
	if err != nil {
		return PushResponse{}, err
	}
	query := target.Query()
	query.Set("msgType", req.MsgType)
	if req.MsgType == "html" && req.HTMLHeight > 0 {
		query.Set("htmlHeight", strconv.Itoa(req.HTMLHeight))
	}
	target.RawQuery = query.Encode()

	body := map[string]string{
		"title": req.Title,
		"msg":   req.Msg,
	}
	if req.URL != "" {
		body["url"] = req.URL
	}
	if req.ImgURL != "" {
		body["imgUrl"] = req.ImgURL
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return PushResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(encoded))
	if err != nil {
		return PushResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return PushResponse{}, err
	}
	defer httpResp.Body.Close()

	respBody, readErr := io.ReadAll(io.LimitReader(httpResp.Body, maxResponseBytes))
	resp := PushResponse{StatusCode: httpResp.StatusCode, Body: string(respBody)}
	if readErr != nil {
		return resp, readErr
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return resp, fmt.Errorf("meow upstream returned %d", httpResp.StatusCode)
	}
	return resp, nil
}

func (c *Client) PushWithRetry(ctx context.Context, req PushRequest) (PushResponse, int, error) {
	maxRetries := 3
	delays := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastResp PushResponse
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		resp, err := c.Push(ctx, req)
		lastResp = resp
		lastErr = err

		// 成功或客户端错误（4xx）不重试
		if err == nil && resp.StatusCode < 500 {
			return resp, i, nil
		}

		// 最后一次不等待
		if i < maxRetries {
			select {
			case <-ctx.Done():
				return lastResp, i, ctx.Err()
			case <-time.After(delays[i]):
			}
		}
	}

	return lastResp, maxRetries, lastErr
}
