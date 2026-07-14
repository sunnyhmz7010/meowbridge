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
	baseURL    string
	httpClient *http.Client
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
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Push(ctx context.Context, req PushRequest) (PushResponse, error) {
	target, err := url.Parse(c.baseURL + "/" + url.PathEscape(req.Nickname))
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
