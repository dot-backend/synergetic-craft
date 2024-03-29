package clienthttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClientAPI interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientHTTP interface {
	Do(ctx context.Context, req *http.Request) (response []byte, statusCode int, err error)
	DoWithTimeout(ctx context.Context, req *http.Request, timeout int64, expectedCode int, out interface{}) error
}

type clientHttp struct {
	domain string
	client ClientAPI
}

func NewClientHTTP(clientAPI ClientAPI, domain string) ClientHTTP {
	return &clientHttp{
		domain: domain,
		client: clientAPI,
	}
}

func (c *clientHttp) Do(ctx context.Context, req *http.Request) (response []byte, statusCode int, err error) {
	return c.do(ctx, req, 0)
}

func (c *clientHttp) DoWithTimeout(ctx context.Context, req *http.Request, timeout int64, expectedCode int, out interface{}) error {
	resp, statusCode, err := c.do(ctx, req, timeout)
	if err != nil {
		return err
	}

	if statusCode != expectedCode {
		return fmt.Errorf("status Code [ %d ], err: %v", statusCode, err)
	}

	err = json.Unmarshal(resp, &out)
	if err != nil {
		return err
	}

	return nil
}

func (c *clientHttp) do(ctx context.Context, req *http.Request, timeout int64) (response []byte, statusCode int, err error) {
	if timeout > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()

		ctx = ctxWithTimeout
	}

	url := c.domain + req.URL.Path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}

	req.URL, err = req.URL.Parse(url)
	if err != nil {
		return
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return body, resp.StatusCode, nil
}
