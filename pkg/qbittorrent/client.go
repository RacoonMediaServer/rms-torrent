package qbittorrent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	torrentutils "github.com/RacoonMediaServer/rms-torrent/pkg/torrent-utils"
)

const apiBasePath = "/api/v2"

type Client struct {
	URL string
	sid string
}

func (c *Client) Login(user, password string) error {
	u, err := url.JoinPath(c.URL, apiBasePath, "auth", "login")
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Add("username", user)
	form.Add("password", password)

	resp, err := http.PostForm(u, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	cookies := resp.Cookies()
	if len(cookies) != 0 {
		c.sid = cookies[0].Raw
	}

	return nil
}

func (c *Client) GetTorrents(ctx context.Context) (result []TorrentInfo, err error) {
	req := apiRequest{
		method:     http.MethodGet,
		apiName:    "torrents",
		methodName: "info",
	}
	err = c.do(ctx, &req, &result)
	return
}

func (c *Client) GetTorrent(ctx context.Context, hash string) (result []TorrentInfo, err error) {
	req := apiRequest{
		method:     http.MethodGet,
		apiName:    "torrents",
		methodName: "info",
		query:      url.Values{},
	}
	req.query.Add("hashes", hash)

	err = c.do(ctx, &req, &result)
	return
}

func (c *Client) SetPreferences(ctx context.Context, p Preferences) error {
	form := url.Values{}
	data, err := json.Marshal(&p)
	if err != nil {
		return err
	}

	form.Add("json", string(data))
	req := apiRequest{
		method:     http.MethodPost,
		apiName:    "app",
		methodName: "setPreferences",
		form:       form,
	}
	return c.do(ctx, &req, nil)
}

func (c *Client) RemoveTorrent(ctx context.Context, id string) error {
	query := url.Values{}
	query.Add("hashes", id)
	query.Add("deleteFiles", "true")

	req := apiRequest{
		method:     http.MethodPost,
		apiName:    "torrents",
		methodName: "delete",
		form:       query,
	}
	return c.do(ctx, &req, nil)
}

func (c *Client) AddTorrent(ctx context.Context, category, tags string, location *string, content []byte) error {
	buf := bytes.NewBuffer([]byte{})
	writer := multipart.NewWriter(buf)
	defer writer.Close()
	if !torrentutils.IsMagnetLink(content) {
		w, err := writer.CreateFormFile("torrents", "download.torrent")
		if err != nil {
			return fmt.Errorf("serialize file failed: %w", err)
		}
		_, err = io.Copy(w, bytes.NewReader(content))
		if err != nil {
			return fmt.Errorf("copy file failed: %w", err)
		}
	} else {
		if err := writer.WriteField("urls", string(content)); err != nil {
			return fmt.Errorf("write magnet link failed: %w", err)
		}
	}
	_ = writer.WriteField("category", category)
	_ = writer.WriteField("sequentialDownload", "true")
	_ = writer.WriteField("firstLastPiecePrio", "true")
	_ = writer.WriteField("tags", tags)
	if location != nil {
		_ = writer.WriteField("savepath", *location)
	}

	req := apiRequest{
		method:      http.MethodPost,
		apiName:     "torrents",
		methodName:  "add",
		body:        bytes.NewReader(buf.Bytes()),
		contentType: writer.FormDataContentType(),
	}
	return c.do(ctx, &req, nil)
}

func (c *Client) GetFiles(ctx context.Context, hash string) (files []File, err error) {
	req := apiRequest{
		method:     http.MethodGet,
		apiName:    "torrents",
		methodName: "files",
		query:      url.Values{},
	}
	req.query.Add("hash", hash)

	err = c.do(ctx, &req, &files)
	return
}

func (c *Client) IncPriority(ctx context.Context, hash string) error {
	req := apiRequest{
		method:     http.MethodPost,
		apiName:    "torrents",
		methodName: "increasePrio",
		form:       url.Values{},
	}
	req.form.Add("hashes", hash)
	return c.do(ctx, &req, nil)
}
