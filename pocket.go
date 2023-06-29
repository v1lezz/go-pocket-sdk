package go_pocket_sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	host                 = "https://getpocket.com/v3"
	authorizeUrl         = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"
	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"
	xErrorHeader         = "X-Error"
	defaultTimeout       = 5 * time.Second
)

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirect_uri"`
	}
	authorizeRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}
	AuthorizeResponse struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}
	addRequest struct {
		URL         string `json:"url"`
		Title       string `json:"title,omitempty"`
		Tags        string `json:"tags,omitempty"`
		AccessToken string `json:"access_token"`
		ConsumerKey string `json:"consumer_key"`
	}

	AddInput struct {
		URL         string
		Title       string
		Tags        []string
		AccessToken string
	}
)

func (i AddInput) validate() error {
	if i.URL == "" {
		return errors.New("required URL values is empty")
	}
	if i.AccessToken == "" {
		return errors.New("empty access token")
	}
	return nil
}

func (i AddInput) generate(consumerKey string) addRequest {
	return addRequest{
		URL:         i.URL,
		Title:       i.Title,
		Tags:        strings.Join(i.Tags, ","),
		AccessToken: i.AccessToken,
		ConsumerKey: consumerKey,
	}
}

type Client struct {
	client      *http.Client
	consumerKey string
}

func NewClient(consumerKey string) (*Client, error) {
	if consumerKey == "" {
		return nil, errors.New("consumer k	ey is empty")
	}
	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		consumerKey: consumerKey,
	}, nil
}

func (c *Client) GetRequestToken(ctx context.Context, redirectUrl string) (string, error) {
	inp := &requestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUrl,
	}
	values, err := c.doHTTP(ctx, endpointRequestToken, inp)
	if err != nil {
		return "", err
	}
	if values.Get("code") == "" {
		return "", errors.New("empty request token in API response")
	}
	return values.Get("code"), nil
}

func (c *Client) GetAuthorizationURL(requestToken, redirectURL string) (string, error) {
	if requestToken == "" || redirectURL == "" {
		return "", errors.New("empty params")
	}
	return fmt.Sprintf(authorizeUrl, requestToken, redirectURL), nil
}

func (c *Client) Authorize(ctx context.Context, requestToken string) (*AuthorizeResponse, error) {
	if requestToken == "" {
		return nil, errors.New("empty request token")
	}

	inp := &authorizeRequest{
		Code:        requestToken,
		ConsumerKey: c.consumerKey,
	}
	values, err := c.doHTTP(ctx, endpointAuthorize, inp)
	if err != nil {
		return nil, err
	}
	accessToken, username := values.Get("access_token"), values.Get("username")
	if accessToken == "" {
		return nil, errors.New("empty access token in API response")
	}
	return &AuthorizeResponse{
		AccessToken: accessToken,
		Username:    username,
	}, nil
}

func (c *Client) doHTTP(ctx context.Context, endpoint string, body interface{}) (url.Values, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to marshal input body")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to create new request")
	}
	req.Header.Set("Content-Type", "application-json; charset=UTF8")
	resp, err := c.client.Do(req)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to send http request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("API Error: %s", resp.Header.Get(xErrorHeader))
		return url.Values{}, errors.New(err)
	}
	respB, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to read request body")
	}
	values, err := url.ParseQuery(string(respB))
	if err != nil {
		return url.Values{}, errors.WithMessage(err, "failed to parse response body")
	}
	return values, nil
}
