package utils

import (
	"context"
	"crypto/tls"
	"net/http"

	"golang.org/x/crypto/acme"
)

func CheckAcmeServer(rawUrl string) error {
	// Create a new HTTP POST request to the webhook URL.

	// if rawUrl == "" {
	// 	rawUrl = "https://127.0.0.1:14000/dir"
	// }

	client := &acme.Client{
		DirectoryURL: rawUrl,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}

	_, err := client.Discover(context.Background())
	return err
}
