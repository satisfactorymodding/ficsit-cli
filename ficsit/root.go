package ficsit

import (
	"net/http"

	"github.com/Khan/genqlient/graphql"
	"github.com/spf13/viper"
)

func InitAPI() graphql.Client {
	return graphql.NewClient(viper.GetString("api-base")+viper.GetString("graphql-api"), http.DefaultClient)
}
