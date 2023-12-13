package provider

import (
	"context"
	"github.com/Khan/genqlient/graphql"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type ficsitProvider struct {
	client graphql.Client
}

func initFicsitProvider(client graphql.Client) ficsitProvider {
	return ficsitProvider{
		client,
	}
}

func (p ficsitProvider) Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	return ficsit.Mods(context, p.client, filter)
}

func (p ficsitProvider) GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error) {
	return ficsit.GetMod(context, p.client, modReference)
}

func (p ficsitProvider) ModVersions(context context.Context, modReference string, filter ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error) {
	return ficsit.ModVersions(context, p.client, modReference, filter)
}

func (p ficsitProvider) SMLVersions(context context.Context) (*ficsit.SMLVersionsResponse, error) {
	return ficsit.SMLVersions(context, p.client)
}

func (p ficsitProvider) ModVersionsWithDependencies(_ context.Context, modID string) (*ficsit.AllVersionsResponse, error) {
	return ficsit.GetAllModVersions(modID)
}

func (p ficsitProvider) GetModName(context context.Context, modReference string) (*ficsit.GetModNameResponse, error) {
	return ficsit.GetModName(context, p.client, modReference)
}
