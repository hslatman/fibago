package client

import (
	"encoding/json"

	"github.com/gregjones/httpcache"
	"github.com/sendgrid/rest"
)

const version = "0.0.1"

var (
	BaseURL             = "https://api.fingerbank.org"
	UserAgent           = "https://github.com/hslatman/fibago"
	EndpointInterrogate = "/api/v2/combinations/interrogate"
)

type Client struct {
	baseURL string
	apiKey  string
	Cache   httpcache.Cache
}

func NewClient(apiKey string) (*Client, error) {

	client := &Client{
		baseURL: BaseURL,
		apiKey:  apiKey,
	}

	return client, nil
}

func (c *Client) Interrogate(fingerprint string) (*rest.Response, error) {

	// Build the URL
	url := c.baseURL + EndpointInterrogate

	// Build the request headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = UserAgent

	// GET Combinations
	method := rest.Get

	// Build the query parameters
	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey

	// Build the body
	body := make(map[string]string)
	body["dhcp_fingerprint"] = fingerprint
	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the API call
	request := rest.Request{
		Method:      method,
		BaseURL:     url,
		Headers:     headers,
		QueryParams: queryParams,
		Body:        requestBody,
	}

	response, err := c.checkCache(request)
	if err != nil {
		return nil, err
	}

	if response != nil {
		return response, nil
	}

	response, err = rest.Send(request)

	c.updateCache(request, response)

	return response, err
}
