package fibago

import (
	//"fmt"

	"strconv"
	"strings"

	"github.com/sendgrid/rest"
)

const (
	version     = "0.0.1"
	baseURL     = "https://api.fingerbank.org/api/v2"
	userAgent   = "https://github.com/hslatman/fibago"
	cacheHeader = "X-From-Cache"

	endpointInterrogate     = "/combinations/interrogate"
	endpointDevices         = "/devices/"
	endpointDevicesBaseInfo = "/devices/base_info"
	endpointOUI             = "/oui"
	endpointStatic          = "/download/db"
	endpointUsers           = "/users"
)

type ClientModifier func(c *Client)

type Client struct {
	baseURL     string
	apiKey      string
	userAgent   string
	modifiers   []ClientModifier
	logger      Logger
	cache       Cache
	cacheHeader string
}

func NewClient(apiKey string, modifiers ...ClientModifier) *Client {

	c := &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		userAgent:   userAgent,
		cacheHeader: cacheHeader,
	}

	c.modifiers = append(c.modifiers, modifiers...)
	for _, modifier := range c.modifiers {
		modifier(c)
	}

	c.debug("configured client") // TODO: some way to introspect the modifiers that were applied and print it nicely?

	return c
}

func WithLogger(logger Logger) ClientModifier {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithBaseURL(baseURL string) ClientModifier {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithUserAgent(userAgent string) ClientModifier {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

type InterrogateParameters struct {
	DHCPFingerprint string
	UserAgents      []string
	MACAddress      string
}

func (c *Client) Interrogate(params *InterrogateParameters) (*rest.Response, error) {

	// TODO: probably want to refactor some of the requests into a more generic function to use
	// TODO: do we want to keep InterrogateParameters like this, or provide some other way for the parameters to be passed, like the modifier approach?

	url := c.baseURL + endpointInterrogate

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = c.userAgent

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

	c.debug("preparing request")

	request := rest.Request{
		Method:      rest.Get,
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

	c.debug("executing request")

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

	url := c.baseURL + endpointDevices //+ strconv.Itoa(id)

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = c.userAgent

	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey
	queryParams["id"] = strconv.Itoa(id)

	request := rest.Request{
		Method:      rest.Get,
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

func (c *Client) DeviceIsA(id string, otherID string) {
	// TODO: implement
}

func (c *Client) DevicesBaseInfo() (*rest.Response, error) {

	url := c.baseURL + endpointDevicesBaseInfo

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = c.userAgent

	queryParams := make(map[string]string)
	queryParams["key"] = c.apiKey

	// TODO add fields handling Comma delimited list of fields to have in the dump. Allowed fields are: id, name, parent_id, virtual_parent_id, details. Default value is ‘id,name’ when the parameter isn’t specified
	// This call results in a large JSON, like 3, 4 MB, which may get bigger with the other fields. We probably want to order the fields in such a way that the response can be
	// cached more deterministically

	request := rest.Request{
		Method:      rest.Get,
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

	url := c.baseURL + endpointUsers //+ "/" + c.apiKey // TODO: this seems to be incorrect? devices in path is incorrect too. Documentation on Fingerbank website is a bit weird too.

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["User-Agent"] = c.userAgent

	// Build the query parameters
	//queryParams := make(map[string]string)
	//queryParams["key"] = c.apiKey
	//queryParams["account_key"] = c.apiKey

	request := rest.Request{
		Method:  rest.Get,
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
