package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	key           = "msg"
	retryDuration = 2 * time.Second
)

func main() {
	var (
		accountID, namespaceID, apiToken string
		exists                           bool
	)

	accountID, exists = os.LookupEnv("CLOUDFLARE_ACCOUNT_ID")
	if !exists {
		panic("set CLOUDFLARE_ACCOUNT_ID env")
	}
	namespaceID, exists = os.LookupEnv("CLOUDFLARE_NAMESPACE_ID")
	if !exists {
		panic("set CLOUDFLARE_NAMESPACE_ID env")
	}
	apiToken, exists = os.LookupEnv("CLOUDFLARE_API_TOKEN")
	if !exists {
		panic("set CLOUDFLARE_API_TOKEN env")
	}

	t := time.NewTicker(retryDuration)
	defer t.Stop()

	var value string
	for {
		v := getAndDeleteKey(key, accountID, namespaceID, apiToken)
		if len(v) > 0 {
			value = v
			break
		}
		<-t.C
		log.Println("Key not found, retrying...")
	}

	// TODO: invoke halp.
	fmt.Println("Value:", value)
}

func getAndDeleteKey(key, accountID, namespaceID, apiToken string) string {
	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/storage/kv/namespaces/%s/values/%s",
		accountID, namespaceID, key,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "" // blank string indicates retry
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}

	value, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	deleteKey(key, accountID, namespaceID, apiToken)

	return string(value)
}

func deleteKey(key, accountID, namespaceID, apiToken string) {
	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/storage/kv/namespaces/%s/values/%s",
		accountID, namespaceID, key,
	)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %s\n", string(body))
		os.Exit(1)
	}
}
