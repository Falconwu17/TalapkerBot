package contentclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	Base string
	HC   *http.Client
}

type Content struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func New(base string) *Client {
	return &Client{
		Base: base,
		HC:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Get(slug, lang string) (Content, error) {
	var out Content
	u := fmt.Sprintf("%s/content?slug=%s&lang=%s", c.Base, url.QueryEscape(slug), url.QueryEscape(lang))
	resp, err := c.HC.Get(u)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return out, fmt.Errorf("status %d", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&out)
	return out, err
}
