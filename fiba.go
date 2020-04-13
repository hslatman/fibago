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
