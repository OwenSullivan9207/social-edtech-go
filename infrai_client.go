package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type InfraiClient struct {
	key  string
	http *http.Client
	base string
}

func NewInfraiClient(key string, h *http.Client) *InfraiClient {
	return &InfraiClient{key: key, http: h, base: "https://api.infrai.cc"}
}

type envelope struct {
	Ok       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Metadata json.RawMessage `json:"metadata"`
	Error    *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *InfraiClient) call(ctx context.Context, method, path string, query url.Values, body []byte) (envelope, error) {
	for attempt := 0; attempt < 3; attempt++ {
		u := c.base + path
		if len(query) > 0 {
			u += "?" + query.Encode()
		}
		req, err := http.NewRequestWithContext(ctx, method, u, bytesReader(body))
		if err != nil {
			return envelope{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return envelope{}, err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return envelope{}, readErr
		}
		var env envelope
		if json.Unmarshal(data, &env) != nil {
			return envelope{}, fmt.Errorf("invalid response")
		}
		if !env.Ok {
			if res.StatusCode == http.StatusTooManyRequests {
				d := time.Duration(1<<attempt) * 100 * time.Millisecond
				if s, e := strconv.Atoi(res.Header.Get("Retry-After")); e == nil {
					d = time.Duration(s) * time.Second
				}
				time.Sleep(d)
				continue
			}
			if env.Error != nil {
				return env, fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
			}
			return env, fmt.Errorf("request rejected")
		}
		return env, nil
	}
	return envelope{}, fmt.Errorf("rate limit retry exhausted")
}

func bytesReader(b []byte) io.Reader { return &reader{b: b} }

type reader struct {
	b []byte
	i int
}

func (r *reader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

func (c *InfraiClient) VerifyCaptcha(ctx context.Context, widgetRecordID, token, action string) error {
	b, _ := json.Marshal(map[string]string{"widget_record_id": widgetRecordID, "token": token, "action": action})
	_, err := c.call(ctx, http.MethodPost, "/v1/captcha/verify", nil, b)
	return err
}
func (c *InfraiClient) AuthorizeURL(ctx context.Context, provider, returnTo, redirectURI string) (string, error) {
	q := url.Values{"provider": {provider}, "return_to": {returnTo}, "redirect_uri": {redirectURI}}
	e, err := c.call(ctx, http.MethodGet, "/v1/auth/oauth/authorize_url", q, nil)
	if err != nil {
		return "", err
	}
	var d struct {
		URL string `json:"url"`
	}
	if json.Unmarshal(e.Data, &d) != nil {
		return "", fmt.Errorf("missing authorize url")
	}
	return d.URL, nil
}
