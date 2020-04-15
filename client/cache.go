package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gregjones/httpcache"
	"github.com/sendgrid/rest"
)

type CacheModifier func(c *Client) // TODO: we could make it work on Cache, but then some more work is required to make it work with httpcache.Cache (or do that part ourselves)

type Cache interface {
	httpcache.Cache
}

// TODO: add "cache backend" specific implementations, like httpcache/diskcache, in subpackages

func WithCache(cache Cache, modifiers ...CacheModifier) ClientModifier {
	return func(c *Client) {
		c.cache = cache
		for _, modifier := range modifiers {
			modifier(c)
		}
	}
}

func WithCacheHeader(cacheHeader string) CacheModifier {
	return func(c *Client) {
		c.cacheHeader = cacheHeader
	}
}

func (c *Client) checkCache(r rest.Request) (*rest.Response, error) {

	c.debug("checking cache")

	if c.cache == nil {
		c.debug("no cache configured")
		return nil, nil
	}

	ck, err := cacheKey(r)
	if err != nil {
		c.debug(fmt.Sprintf("error creating cache key: %s", err.Error()))
		return nil, err
	}

	c.debug(fmt.Sprintf("looking up key: %s", ck))

	cachedResponse, ok := c.cache.Get(ck)
	if ok {
		var response rest.Response
		err := json.Unmarshal(cachedResponse, &response)
		if c.cacheHeader != "" {
			response.Headers[c.cacheHeader] = []string{"1"}
		}

		c.debug("returning cached response")
		return &response, err
	}

	c.debug("no cached response found")
	// No cached response found; return empty
	return nil, nil
}

func (c *Client) updateCache(req rest.Request, resp *rest.Response) error {

	c.debug("updating cache")

	if c.cache == nil {
		c.debug("no cache configured")
		return nil
	}

	if resp == nil {
		return nil
	}

	if resp.StatusCode != 200 {
		c.debug(fmt.Sprintf("abort storing response with status code %d", resp.StatusCode))
		return nil
	}

	ck, err := cacheKey(req)
	if err != nil {
		c.debug(fmt.Sprintf("error creating cache key: %s", err.Error()))
		return err
	}

	c.debug(fmt.Sprintf("storing response for ck: %s", ck))

	data, err := json.Marshal(resp) // TODO: we may want to store this as an http.Request; not a rest.Request
	if err != nil {
		c.debug(fmt.Sprintf("error serializating response: %s", err.Error()))
		return err
	}

	c.cache.Set(ck, data)

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
		key := r.BaseURL + string(data[:]) // NOTE: all query parameters are used in creating the cache key; alternatively, when using the body, that should be included, instead
		return key, nil
	}

	return "", errors.New("methods other than GET not supported in cache")

}
