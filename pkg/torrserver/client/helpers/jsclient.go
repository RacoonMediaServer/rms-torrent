package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type JsClient struct {
	URL string
}

type TorrentInfo struct {
	Title string
	Hash  string
	Size  uint64 `json:"torrent_size"`
}

func (c *JsClient) GetTorrents(ctx context.Context) ([]TorrentInfo, error) {
	js := map[string]string{
		"action": "list",
	}
	jsBytes, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	req := apiRequest{
		method:     http.MethodPost,
		methodName: "torrents",
		body:       bytes.NewReader(jsBytes),
	}
	result := []TorrentInfo{}
	err = c.do(ctx, &req, &result)

	return result, err
}

func (c *JsClient) GetTorrent(ctx context.Context, hash string) (*TorrentInfo, error) {
	js := map[string]string{
		"action": "get",
		"hash":   hash,
	}
	jsBytes, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	req := apiRequest{
		method:     http.MethodPost,
		methodName: "torrents",
		body:       bytes.NewReader(jsBytes),
	}
	result := TorrentInfo{}
	err = c.do(ctx, &req, &result)

	return &result, err
}

func (c *JsClient) RemoveTorrent(ctx context.Context, hash string) error {
	js := map[string]string{
		"action": "rem",
		"hash":   hash,
	}
	jsBytes, err := json.Marshal(js)
	if err != nil {
		return err
	}
	req := apiRequest{
		method:     http.MethodPost,
		methodName: "torrents",
		body:       bytes.NewReader(jsBytes),
	}

	err = c.do(ctx, &req, nil)
	return err
}
