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

# Notes

This Fingerbank client library is small and imcomplete at the moment.
More functions and API endpoints will be added soon.
Some typed response objects may be added in the future too.

The library uses [sendgrid/rest](https://github.com/sendgrid/rest) under the hood.
