package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type apiRequest struct {
	method      string
	apiName     string
	methodName  string
	form        url.Values
	query       url.Values
	body        io.Reader
	contentType string
}

func (c *Client) do(ctx context.Context, r *apiRequest, dest any) error {
	u, err := url.JoinPath(c.URL, apiBasePath, r.apiName, r.methodName)
	if err != nil {
		return fmt.Errorf("join url failed: %w", err)
	}

	if r.query != nil {
		u = u + "?" + r.query.Encode()
	}

	var body io.Reader
	if r.form != nil {
		body = strings.NewReader(r.form.Encode())
	} else if r.body != nil {
		body = r.body
	}

	req, err := http.NewRequestWithContext(ctx, r.method, u, body)
	if err != nil {
		return fmt.Errorf("compose req failed: %w", err)
	}

	if c.sid != "" {
		req.Header.Add("Cookie", c.sid)
	}
	if r.form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if r.contentType != "" {
		req.Header.Set("Content-Type", r.contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
	}

	if dest != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body failed: %w", err)
		}

		if err = json.Unmarshal(data, dest); err != nil {
			return fmt.Errorf("parse response failed: %w", err)
		}
	}

	return nil
}
