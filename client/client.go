package client

import (
	//"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gregjones/httpcache"
	"github.com/sendgrid/rest"
)

const version = "0.0.1"

var (
	BaseURL                 = "https://api.fingerbank.org"
	UserAgent               = "https://github.com/hslatman/fibago"
	EndpointInterrogate     = "/api/v2/combinations/interrogate"
	EndpointDevices         = "/api/v2/devices/"
	EndpointDevicesBaseInfo = "/api/v2/devices/base_info"
	EndpointOUI             = "/api/v2/oui"
	EndpointStatic          = "/api/v2/download/db"
	EndpointUsers           = "/api/v2/users"
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

func NewClientWithCache(apiKey string, cache httpcache.Cache) (*Client, error) {

	client := &Client{
		baseURL: BaseURL,
		apiKey:  apiKey,
		Cache:   cache,
	}

	return client, nil
}

type InterrogateParameters struct {
	DHCPFingerprint string
	UserAgents      []string
	MACAddress      string
}

func (c *Client) Interrogate(params *InterrogateParameters) (*rest.Response, error) {

	// Build the URL
	url := c.baseURL + EndpointInterrogate

	// Build the request headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = UserAgent

	// GET Combinations
	method := rest.Get

	// Build the query parameters; Fingerbank does not only support querying using the body, but also in query parameters.
	// This is more like normal queries and also works with our caching, so we've changed the implementation to reflect that.
	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey
	if params.DHCPFingerprint != "" {
		queryParams["dhcp_fingerprint"] = params.DHCPFingerprint // TODO: ensure the fingerprint does not contain spaces
	}
	if params.MACAddress != "" {
		queryParams["mac"] = params.MACAddress // TODO: ensure the MAC is shortened (no colons) and lowercase
	}
	if params.UserAgents != nil {
		queryParams["user_agents"] = strings.Join(params.UserAgents, ",") // TODO: documentation looks to allow multiple user agents; this should do it, right?
	}

	fmt.Println(fmt.Sprintf("%+v", queryParams))

	// Build the body; NOTE: this is thus not required anymore, but is good to know that it's possible too.
	// body := make(map[string]string)
	// body["dhcp_fingerprint"] = fingerprint
	// requestBody, err := json.Marshal(body)
	// if err != nil {
	// 	return nil, err
	// }

	// Make the API call
	request := rest.Request{
		Method:      method,
		BaseURL:     url,
		Headers:     headers,
		QueryParams: queryParams,
		//Body:        requestBody,
	}

	fmt.Println(fmt.Sprintf("%+v", request))

	response, err := c.checkCache(request)
	if err != nil {
		return nil, err
	}

	if response != nil {
		return response, nil
	}

	fmt.Println(fmt.Sprintf("%+v", request))

	response, err = rest.Send(request)

	c.updateCache(request, response)

	return response, err
}

func (c *Client) Static() error {
	// TODO: this can be used to download an sqlite3 database with the data. It's not very REST-ish,
	// so I don't think we should use the sendgrid/rest library to do this, but do some streaming file download
	// (in the background?). We should probably implement some way for registering the download as
	// happened before and mark it in the cache, still. The database seems to contain data with latest
	// updates for the day of download (or just before that), so relatively fresh data, but quite
	// a big file to download (almost 600MB), which is not something to do just every time. A download using
	// curl looks like this:
	//
	// $ curl -X GET https://api.fingerbank.org/api/v2/download/db?key=your_api_key -o fingerbank.sqlite
	//
	// This will result in the database being available in the fingerbank.sqlite file, after the download
	// completes.

	return nil
}

func (c *Client) Devices(id int) (*rest.Response, error) {

	// TODO: this one does not seem to work either ...

	url := c.baseURL + EndpointDevices //+ strconv.Itoa(id)

	// Build the request headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = UserAgent

	// GET Combinations
	method := rest.Get

	// Build the query parameters
	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey
	queryParams["id"] = strconv.Itoa(id)

	// Make the API call
	request := rest.Request{
		Method:      method,
		BaseURL:     url,
		Headers:     headers,
		QueryParams: queryParams,
	}

	fmt.Println(fmt.Sprintf("%+v", request))

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

func (c *Client) DeviceIsA(id string, otherID string) {

}

func (c *Client) DevicesBaseInfo() (*rest.Response, error) {
	// Build the URL
	url := c.baseURL + EndpointDevicesBaseInfo

	// Build the request headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = UserAgent

	// GET Combinations
	method := rest.Get

	// Build the query parameters
	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey

	// TODO add fields handling Comma delimited list of fields to have in the dump. Allowed fields are: id, name, parent_id, virtual_parent_id, details. Default value is ‘id,name’ when the parameter isn’t specified
	// This call results in a large JSON, like 3, 4 MB, which may get bigger with the other fields. We probably want to order the fields in such a way that the response can be
	// cached more deterministically

	// Make the API call
	request := rest.Request{
		Method:      method,
		BaseURL:     url,
		Headers:     headers,
		QueryParams: queryParams,
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

func (c *Client) AccountInfo() (*rest.Response, error) {

	// Build the URL
	url := c.baseURL + EndpointUsers //+ "/" + c.apiKey // TODO: this seems to be incorrect? devices in path is incorrect too. Documentation on Fingerbank website is a bit weird too.

	// Build the request headers
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = UserAgent

	// GET Combinations
	method := rest.Get

	// Build the query parameters
	//queryParams := make(map[string]string)
	//queryParams["key"] = c.apiKey
	//queryParams["account_key"] = c.apiKey

	// Make the API call
	request := rest.Request{
		Method:  method,
		BaseURL: url,
		Headers: headers,
		//QueryParams: queryParams,
	}

	// response, err := c.checkCache(request)
	// if err != nil {
	// 	return nil, err
	// }

	// if response != nil {
	// 	return response, nil
	// }

	response, err := rest.Send(request)

	//c.updateCache(request, response)

	return response, err

}
