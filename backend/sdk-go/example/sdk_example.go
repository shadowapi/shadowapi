package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/shadowapi/shadowapi/backend/pkg/api"
)

// MySecuritySource implements the SecuritySource interface.
type MySecuritySource struct {
	token string
}

// BearerAuth returns a BearerAuth instance with the token.
func (s MySecuritySource) BearerAuth(ctx context.Context, operationName api.OperationName) (api.BearerAuth, error) {
	return api.BearerAuth{Token: s.token}, nil
}

// ZitadelCookieAuth returns a zero value; update if needed.
func (s MySecuritySource) ZitadelCookieAuth(ctx context.Context, operationName api.OperationName) (api.ZitadelCookieAuth, error) {
	return api.ZitadelCookieAuth{}, nil
}

// PlainCookieAuth returns a zero value; update if needed.
func (s MySecuritySource) PlainCookieAuth(ctx context.Context, operationName api.OperationName) (api.PlainCookieAuth, error) {
	return api.PlainCookieAuth{}, nil
}

func main() {
	// Parse the server URL.
	// use the same prefix "/api/v1" as in backend/internal/server/server.go
	baseURL, err := url.Parse("https://example.com/api/v1")
	if err != nil {
		log.Fatal("Invalid URL:", err)
	}

	// Provide a security source using BearerAuth with your API key.
	sec := MySecuritySource{token: "XXXX"}

	// Create a new API client.
	// If authentication is needed, implement the SecuritySource interface and pass it here instead of nil.
	client, err := api.NewClient(baseURL.String(), sec)
	if err != nil {
		log.Fatal("Error creating client:", err)
	}

	// Set up a context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Example: Call the DatasourceEmailList endpoint.
	listParams := api.DatasourceEmailListParams{
		Offset: api.NewOptInt32(0),
		Limit:  api.NewOptInt32(10),
	}
	emails, err := client.DatasourceEmailList(ctx, listParams)
	if err != nil {
		log.Fatal("Error listing datasource emails:", err)
	}
	log.Printf("Datasource Emails: %+v", emails)
}
