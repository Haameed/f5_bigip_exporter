package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Haameed/f5_bigip_exporter/internal/utils"
)

type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type F5TokenClient struct {
	tgt   url.URL
	hc    Client
	ctx   context.Context
	token utils.TokenDetails
}

func (c *F5TokenClient) newGetRequest(url string) (*http.Request, error) {
	r, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("X-F5-Auth-Token", c.token.Token)
	return r, nil
}

func (c *F5TokenClient) Get(path string, obj any) error {
	u := c.tgt
	u.Path = path

	req, err := c.newGetRequest(u.String())
	if err != nil {
		return err
	}

	req = req.WithContext(c.ctx)
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code was %d, expected 200, Host: %v , (path: %q)", resp.StatusCode, u.Host, path)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, obj)

}

func (c *F5TokenClient) String() string {
	return c.tgt.String()
}

func newBigIPTokenClient(ctx context.Context, tgt url.URL, hc Client, token utils.TokenDetails) (*F5TokenClient, error) {
	return &F5TokenClient{tgt, hc, ctx, token}, nil
}
