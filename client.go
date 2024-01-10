package kiwi_sdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sllt/af/convertor"
	"time"
)

var ErrInvalidResponse = errors.New("invalid response")

type (
	Client struct {
		client     *resty.Client
		url        string
		authorizer authStore
	}
	ClientOption func(*Client)
)

func NewClient(url string, opts ...ClientOption) *Client {
	client := resty.New()
	client.
		SetRetryCount(3).
		SetRetryWaitTime(3 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second)

	c := &Client{
		client:     client,
		url:        url,
		authorizer: authorizeNoOp{},
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithDebug() ClientOption {
	return func(c *Client) {
		c.client.SetDebug(true)
	}
}

func WithAdminEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/admins/auth-with-password", email, password)
	}
}

func WithUserEmailPassword(email, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/users/auth-with-password", email, password)
	}
}

func WithCollectionAuth(collection, username, password string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeEmailPassword(c.client, c.url+"/api/collections/"+collection+"/auth-with-password", username, password)
	}

}

func WithAdminToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/admins/auth-refresh", token)
	}
}

func WithUserToken(token string) ClientOption {
	return func(c *Client) {
		c.authorizer = newAuthorizeToken(c.client, c.url+"/api/collections/users/auth-refresh", token)
	}
}

func (c *Client) Authorize() error {
	return c.authorizer.authorize()
}

func (c *Client) Update(collection string, id string, body any) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetBody(body)

	resp, err := request.Patch(c.url + "/api/collections/{collection}/records/" + id)
	if err != nil {
		return fmt.Errorf("[update] can't send update request to kiwi, err %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("[update] kiwi returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

func (c *Client) Create(collection string, body any) (ResponseCreate, error) {
	var response ResponseCreate

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetBody(body).
		SetResult(&response)

	resp, err := request.Post(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[create] can't send update request to kiwi, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[create] kiwi returned status: %d, msg: %s, body: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			fmt.Sprintf("%+v", body), // TODO remove that after debugging
			ErrInvalidResponse,
		)
	}

	return *resp.Result().(*ResponseCreate), nil
}

func (c *Client) Delete(collection string, id string) error {
	if err := c.Authorize(); err != nil {
		return err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection).
		SetPathParam("id", id)

	resp, err := request.Delete(c.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return fmt.Errorf("[delete] can't send update request to kiwi, err %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("[delete] kiwi returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	return nil
}

func (c *Client) List(collection string, params *Option) (ResponseList[map[string]any], error) {
	var response ResponseList[map[string]any]

	if err := c.Authorize(); err != nil {
		return response, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection)

	if params != nil {
		if params.Page > 0 {
			request.SetQueryParam("page", convertor.ToString(params.Page))
		}
		if params.Size > 0 {
			request.SetQueryParam("perPage", convertor.ToString(params.Size))
		}
		if params.Filters != "" {
			request.SetQueryParam("filter", params.Filters)
		}
		if params.Sort != "" {
			request.SetQueryParam("sort", params.Sort)
		}
	}
	resp, err := request.Get(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return response, fmt.Errorf("[list] can't send update request to kiwi, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[list] kiwi returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	var responseRef any = &response
	if params != nil && params.hackResponseRef != nil {
		responseRef = params.hackResponseRef
	}
	if err := json.Unmarshal(resp.Body(), responseRef); err != nil {
		return response, fmt.Errorf("[list] can't unmarshal response, err %w", err)
	}
	return response, nil
}

func (c *Client) list(collection string, params Option) (*resty.Response, error) {
	if err := c.Authorize(); err != nil {
		return nil, err
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", collection)

	if params.Page > 0 {
		request.SetQueryParam("page", convertor.ToString(params.Page))
	}
	if params.Size > 0 {
		request.SetQueryParam("perPage", convertor.ToString(params.Size))
	}
	if params.Filters != "" {
		request.SetQueryParam("filter", params.Filters)
	}
	if params.Sort != "" {
		request.SetQueryParam("sort", params.Sort)
	}

	resp, err := request.Get(c.url + "/api/collections/{collection}/records")
	if err != nil {
		return nil, fmt.Errorf("can't send update request to kiwi, err %w", err)
	}

	return resp, nil
}

func (c *Client) AuthStore() authStore {
	return c.authorizer
}
