package fibago

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

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
	if !ok {
		c.debug("no cached response found")
		return nil, nil
	}

	var response rest.Response
	err = json.Unmarshal(cachedResponse, &response)
	if err != nil {
		return nil, err
	}

	date, ok := response.Headers["Date"]
	if !ok {
		// NOTE: this effectively results in the actual request being performed whenever the Date header is not set
		return nil, nil
	}

	// TODO: also implement Max-Age?
	// TODO: does this work properly with the timezone? the actual time that a response is considered valid also depends on the timezone now.
	cachedResponseDate, _ := http.ParseTime(date[0])
	numSeconds := c.cacheTimeInSeconds
	duration := time.Duration(numSeconds) * time.Second
	if time.Since(cachedResponseDate) > duration {
		c.debug(fmt.Sprintf("cached response is too old"))
		c.cache.Delete(ck)
		return nil, nil
	}

	if c.cacheHeader != "" {
		response.Headers[c.cacheHeader] = []string{"1"}
	}

	c.debug("returning cached response")
	return &response, err
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

	_, ok := resp.Headers["Date"]
	if !ok {
		date := []string{(time.Now()).Format(http.TimeFormat)} // TODO: timezone?
		c.debug(fmt.Sprintf("adding date header to response: %s", date[0]))
		resp.Headers["Date"] = date
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
		// TODO: order the fields? we've already taken care of it, sort of, by the ordered if-statements for adding it to the queryParams

		authKey := r.QueryParams["key"]
		delete(r.QueryParams, "key") // We're removing the key from the parameters, such that it does not end up in the cache

		request, err := rest.BuildRequestObject(r)
		if err != nil {
			return "", err
		}

		r.QueryParams["key"] = authKey // We're putting the authentication key back where it was

		key := request.URL.String()

		return key, nil
	}

	return "", errors.New("methods other than GET not supported in cache")

}
