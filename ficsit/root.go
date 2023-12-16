package ficsit

import (
	"fmt"
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/spf13/viper"
)

type AuthedTransport struct {
	Wrapped http.RoundTripper
}

func (t *AuthedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := viper.GetString("api-key")
	if key != "" {
		req.Header.Set("Authorization", key)
	}

	rt, err := t.Wrapped.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("failed roundtrip: %w", err)
	}

	return rt, nil
}

func InitAPI() graphql.Client {
	httpClient := http.Client{
		Transport: &AuthedTransport{
			Wrapped: http.DefaultTransport,
		},
	}

	return graphql.NewClient(viper.GetString("api-base")+viper.GetString("graphql-api"), &httpClient)
}
