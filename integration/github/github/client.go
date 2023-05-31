package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultBaseUTL      = "https://api.github.com/"
	acceptVersionHeader = "application/vnd.github.v3+json"
)

type Client struct {
	client  *http.Client
	baseUrl *url.URL
}

func NewClient(ctx context.Context, token string) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token})

	tc := oauth2.NewClient(ctx, ts)
	return newClient(tc)

}
func newClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{client: httpClient}
	c.SetBaseUrl(defaultBaseUTL)
	return c
}

func (c *Client) SetBaseUrl(urlString string) error {
	baseUrl, err := url.Parse(urlString)
	if err != nil {
		return err
	}
	c.baseUrl = baseUrl
	return nil
}

func (c *Client) NewRequest(method, urlString string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	u := c.baseUrl.ResolveReference(rel)
	buf, err := c.encodeBody(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", acceptVersionHeader)
	return req, nil
}

func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := ctxhttp.Do(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	err = c.decodeResponse(resp.Body, v)
	if err != nil {
		return nil, err
	}
	return resp, err
}

type ErrorResponse struct {
	*http.Response
	Message string `json:"message"`
}

func (er *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v", er.Response.Request.Method, er.Response.Request.URL.Path)
}

func CheckResponse(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: resp}
	data, err := ioutil.ReadAll(resp.Body)
	if err == nil && data == nil {
		json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

func (c *Client) encodeBody(body interface{}) (io.ReadWriter, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		} else {
			return buf, nil
		}
		return buf, err
	}
	return nil, nil
}

func (c *Client) decodeResponse(body io.ReadCloser, v interface{}) error {
	if v != nil {
		err := json.NewDecoder(body).Decode(v)
		return err
	}
	return nil
}
