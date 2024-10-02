package provider

import (
	"context"
	"errors"

	"github.com/Khan/genqlient/graphql"
	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/cli/localregistry"
	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type FicsitProvider struct {
	client graphql.Client
}

func NewFicsitProvider(client graphql.Client) FicsitProvider {
	return FicsitProvider{
		client,
	}
}

func (p FicsitProvider) Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	return ficsit.Mods(context, p.client, filter)
}

func (p FicsitProvider) GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error) {
	return ficsit.GetMod(context, p.client, modReference)
}

func (p FicsitProvider) ModVersionsWithDependencies(_ context.Context, modID string) ([]resolver.ModVersion, error) {
	response, err := ficsit.GetAllModVersions(modID)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, errors.New(response.Error.Message)
	}

	localregistry.Add(modID, response.Data)

	return convertFicsitVersionsToResolver(response.Data), nil
}

func (p FicsitProvider) GetModName(context context.Context, modReference string) (*resolver.ModName, error) {
	response, err := ficsit.GetModName(context, p.client, modReference)
	if err != nil {
		return nil, err
	}

	return &resolver.ModName{
		ID:           response.Mod.Id,
		ModReference: response.Mod.Mod_reference,
		Name:         response.Mod.Name,
	}, nil
}

func (p FicsitProvider) IsOffline() bool {
	return false
}
