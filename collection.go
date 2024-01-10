package kiwi_sdk

import (
	"encoding/json"
	"fmt"
)

func NewCollection[T any](c *Client, name string) *Collection[T] {
	return &Collection[T]{
		Client: c,
		Name:   name,
	}
}

func NewDefaultConnection(c *Client, name string) *Collection[map[string]any] {
	return &Collection[map[string]any]{
		Client: c,
		Name:   name,
	}
}

type Collection[T any] struct {
	Client *Client
	Name   string
}

func (c *Collection[T]) GetList(option *Option) (ResponseList[T], error) {
	var response ResponseList[T]

	if option == nil {
		option = &Option{}
	}

	option.hackResponseRef = &response

	_, err := c.Client.List(c.Name, option)
	return response, err
}

func (c *Collection[T]) GetOne(id string) (T, error) {
	var response T

	if err := c.Client.Authorize(); err != nil {
		return response, err
	}

	request := c.Client.client.R().
		SetHeader("Content-Type", "application/json").
		SetPathParam("collection", c.Name).
		SetPathParam("id", id)

	resp, err := request.Get(c.Client.url + "/api/collections/{collection}/records/{id}")
	if err != nil {
		return response, fmt.Errorf("[one] can't send update request to kiwi, err %w", err)
	}

	if resp.IsError() {
		return response, fmt.Errorf("[one] kiwi returned status: %d, msg: %s, err %w",
			resp.StatusCode(),
			resp.String(),
			ErrInvalidResponse,
		)
	}

	if err := json.Unmarshal(resp.Body(), &response); err != nil {
		return response, fmt.Errorf("[one] can't unmarshal response, err %w", err)
	}
	return response, nil
}

func (c *Collection[T]) Create(data T) (ResponseCreate, error) {
	return c.Client.Create(c.Name, data)
}

func (c *Collection[T]) Update(id string, data T) error {
	return c.Client.Update(c.Name, id, data)
}

func (c *Collection[T]) Delete(id string) error {
	return c.Client.Delete(c.Name, id)
}
