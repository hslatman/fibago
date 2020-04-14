# fibago

A (small and incomplete) Go client for the [Fingerbank API](https://api.fingerbank.org/).

## Usage

Retrieve the fibago library:

```bash
$ go get github.com/hslatman/fibago/client
```

You'll need to get an API key from Fingerbank.

Example looking up a DHCP fingerprint:

```go
package main

import (
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/hslatman/fibago/client"
)

func main() {

	client, err := client.NewClient("<apikey>")
	if err != nil {
		fmt.Println(err)
		return
	}

	response, err := client.Interrogate("1,15,3,6,44,46,47,31,33,121,249,43") // Example DHCP fingerprint

	if err != nil {
		fmt.Println(err)
		return
	}

	status := response.StatusCode
	fmt.Println(fmt.Sprintf("status: %d", status))
	if status == 401 { // API key not set or invalid
		fmt.Println(response.Body)
		return
	}

	if status == 404 { // Query did not result in any result; Unknown device
		fmt.Println(response.Body)
		return
	}

	fmt.Println(response.Body)
	value := gjson.Get(response.Body, "device") // NOTE: example using gjson for extracting values from JSON
	fmt.Println(value)
}
```

Running the example should show output similar to what is shown below:

```bash
$ go run fiba.go
status: 200
{"device":{"can_be_more_precise":true,"child_devices_count":13,"child_virtual_devices_count":5,"created_at":"2014-09-09T15:09:50.000Z","id":1,"name":"Windows OS","parent_id":16879,"parents":[{"created_at":"2017-09-14T18:41:06.000Z","id":16879,"name":"Operating System","parent_id":null,"updated_at":"2020-04-09T06:58:16.000Z","virtual_parent_id":null}],"updated_at":"2020-02-08T07:38:14.000Z","virtual_parent_id":null},"device_name":"Operating System/Windows OS","request_id":"b41dbcb2-11c7-45e3-a08c-6ab72a478c8c","score":87,"version":""}
{"can_be_more_precise":true,"child_devices_count":13,"child_virtual_devices_count":5,"created_at":"2014-09-09T15:09:50.000Z","id":1,"name":"Windows OS","parent_id":16879,"parents":[{"created_at":"2017-09-14T18:41:06.000Z","id":16879,"name":"Operating System","parent_id":null,"updated_at":"2020-04-09T06:58:16.000Z","virtual_parent_id":null}],"updated_at":"2020-02-08T07:38:14.000Z","virtual_parent_id":null}
```

# Caching

The API client can perform basic caching, although its implementation is experimental for now.
Fingerbank allows 300 requests per hour, which is quite nice and permissive, but there may be cases in which this limit is reached before an hour passes, so caching may help preventing this limit from being reached when many similar queries are executed.
Due to the fact that the Fingerbank API does not send caching headers and requests are not really RESTful, using the RFC 7234 compliant [httpcache](https://github.com/gregjones/httpcache) was not really an option, unfortunately.
Despite this, we've started implementation of a basic method for caching that is inspired by and should be largely compatible with said httpcache library.
Currently the [diskcache](https://github.com/gregjones/httpcache/tree/master/diskcache) backed by [diskv](https://github.com/peterbourgon/diskv) is supported.

Caching can be enabled by creating a cache and setting it on the Client:

```go
package main

import (
	"fmt"
	"github.com/hslatman/fibago/client"
)

func main() {

    cache := client.NewDiskCache("./cache") // This will create the cache directory in the current working directory
    
    client, err := client.NewClient("<apikey>")
    if err != nil {
        fmt.Println(err)
        return
    }

    client.Cache = cache // This line enables caching responses from the Fingerbank API
    
    response, err := client.Interrogate("1,15,3,6,44,46,47,31,33,121,249,43") // Example DHCP fingerprint

	if err != nil {
		fmt.Println(err)
		return
    }
    
    fmt.Println(response.Body)
    fmt.Println(response.Headers)
}
```

An example of using the cache looks like this:

```bash
$ go run fiba.go
{"device":{"can_be_more_precise":true,"child_devices_count":13,"child_virtual_devices_count":5,"created_at":"2014-09-09T15:09:50.000Z","id":1,"name":"Windows OS","parent_id":16879,"parents":[{"created_at":"2017-09-14T18:41:06.000Z","id":16879,"name":"Operating System","parent_id":null,"updated_at":"2020-04-09T06:58:16.000Z","virtual_parent_id":null}],"updated_at":"2020-02-08T07:38:14.000Z","virtual_parent_id":null},"device_name":"Operating System/Windows OS","request_id":"cefc1482-7775-43cb-b4fc-f8526f88a6fa","score":87,"version":""}
map[Content-Length:[539] Content-Type:[application/json] Date:[Mon, 13 Apr 2020 14:50:46 GMT] Server:[Caddy Caddy Caddy]]
$ go run fiba.go
{"device":{"can_be_more_precise":true,"child_devices_count":13,"child_virtual_devices_count":5,"created_at":"2014-09-09T15:09:50.000Z","id":1,"name":"Windows OS","parent_id":16879,"parents":[{"created_at":"2017-09-14T18:41:06.000Z","id":16879,"name":"Operating System","parent_id":null,"updated_at":"2020-04-09T06:58:16.000Z","virtual_parent_id":null}],"updated_at":"2020-02-08T07:38:14.000Z","virtual_parent_id":null},"device_name":"Operating System/Windows OS","request_id":"cefc1482-7775-43cb-b4fc-f8526f88a6fa","score":87,"version":""}
map[Content-Length:[539] Content-Type:[application/json] Date:[Mon, 13 Apr 2020 14:50:46 GMT] Server:[Caddy Caddy Caddy] X-From-Cache:[1]]
```

Note the additional header, X-From-Cache, that is set when the response comes from the cache

# Notes

This Fingerbank client library is small and imcomplete at the moment.
More functions and API endpoints will be added soon.
Some typed response objects may be added in the future too.

The library uses [sendgrid/rest](https://github.com/sendgrid/rest) under the hood.

# TODO

* More API endpoints
* Cache invalidation (either based on date and some timeout, or passing some additional parameter from client, or something different?)
* Nicer cache approach with plain http responses (not json marshalled) and extending the default HTTP client (i.e. RoundTripper?)
* Typed responses
* Make the client implementation use a local instance of the Fingerbank data (see Static() function and description on [fingerbank/perl-client](https://github.com/fingerbank/perl-client/blob/master/client-development-guidelines.md))
* Create a diagram of how the Go client works, including local database lookup, online lookups, cached lookups, etc
* Tests
