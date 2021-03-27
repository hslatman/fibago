package main

import (
	"fmt"

	"github.com/gregjones/httpcache/diskcache"

	fingerbank "github.com/hslatman/fingerbank-go"
)

func main() {

	cache := diskcache.New("./cache") // This will create the cache directory in the current working directory

	client := fingerbank.NewClient("<apikey>", fingerbank.WithCache(cache))

	params := &fingerbank.InterrogateParameters{
		DHCPFingerprint: "1,15,3,6,44,46,47,31,33,121,249,43", // Example DHCP fingerprint
	}
	response, err := client.Interrogate(params) // Example DHCP fingerprint

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(response.Body)
	fmt.Println(response.Headers)
}
