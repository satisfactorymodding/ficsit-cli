package ficsit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/viper"
)

const allVersionEndpoint = `/v1/mod/%s/versions/all`

func GetAllModVersions(modID string) (*AllVersionsResponse, error) {
	response, err := http.DefaultClient.Get(viper.GetString("api-base") + fmt.Sprintf(allVersionEndpoint, modID))
	if err != nil {
		return nil, fmt.Errorf("failed fetching all versions: %w", err)
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	allVersions := AllVersionsResponse{}
	if err := json.Unmarshal(body, &allVersions); err != nil {
		return nil, fmt.Errorf("failed parsing json: %w", err)
	}

	return &allVersions, nil
}
