package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gregjones/httpcache/diskcache"
	"github.com/sendgrid/rest"
)

var (
	CacheHeader = "X-From-Cache"
)

func NewDiskCache(basePath string) *diskcache.Cache {
	return diskcache.New(basePath)
}

func (c *Client) checkCache(r rest.Request) (*rest.Response, error) {

	// No cache set; skip cache check
	if c.Cache == nil {
		return nil, nil
	}

	ck, err := cacheKey(r)
	if err != nil {
		return nil, err
	}

	fmt.Println(ck)

	cachedResponse, ok := c.Cache.Get(ck)
	if ok {
		var response rest.Response
		err := json.Unmarshal(cachedResponse, &response)
		response.Headers[CacheHeader] = []string{"1"}
		return &response, err
	}

	// No cached response found; return empty
	return nil, nil
}

func (c *Client) updateCache(req rest.Request, resp *rest.Response) error {

	// Cache is not set; return early
	if c.Cache == nil {
		return nil
	}

	// We can't update the cache when no response available
	if resp == nil {
		return nil
	}

	// Don't store responses that are not OK-ish
	if resp.StatusCode != 200 {
		return nil
	}

	ck, err := cacheKey(req)
	if err != nil {
		return err
	}

	data, err := json.Marshal(resp)
	if err == nil {
		c.Cache.Set(ck, data)
	}

	return nil
}

func cacheKey(r rest.Request) (string, error) {

	if r.Method == rest.Get {
		// TODO: exclude the apikey from the cache key?
		// TODO: order the fields? we've already taken care of it, sort of, by the ordered if-statements for adding it to the queryParams
		data, err := json.Marshal(r.QueryParams)
		if err != nil {
			return "", err
		}
		key := r.BaseURL + string(data[:]) //strings.Join(r.QueryParams, "|") //+ string(r.Body) // NOTE: we include the query in the cache key, because that's how the Fingerbank API works
		return key, nil
	}

	return "", errors.New("methods other than GET not supported in cache")

}
