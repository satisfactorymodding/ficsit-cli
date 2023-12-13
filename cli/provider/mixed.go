package provider

import (
	"context"

	"github.com/Khan/genqlient/graphql"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type MixedProvider struct {
	ficsitProvider ficsitProvider
	localProvider  localProvider
	Offline        bool
}

func InitMixedProvider(client graphql.Client) *MixedProvider {
	return &MixedProvider{
		ficsitProvider: initFicsitProvider(client),
		localProvider:  initLocalProvider(),
		Offline:        false,
	}
}

func (p MixedProvider) Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	if p.Offline {
		return p.localProvider.Mods(context, filter)
	}
	return p.ficsitProvider.Mods(context, filter)
}

func (p MixedProvider) GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error) {
	if p.Offline {
		return p.localProvider.GetMod(context, modReference)
	}
	return p.ficsitProvider.GetMod(context, modReference)
}

func (p MixedProvider) ModVersions(context context.Context, modReference string, filter ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error) {
	if p.Offline {
		return p.localProvider.ModVersions(context, modReference, filter)
	}
	return p.ficsitProvider.ModVersions(context, modReference, filter)
}

func (p MixedProvider) SMLVersions(context context.Context) (*ficsit.SMLVersionsResponse, error) {
	if p.Offline {
		return p.localProvider.SMLVersions(context)
	}
	return p.ficsitProvider.SMLVersions(context)
}

func (p MixedProvider) ModVersionsWithDependencies(context context.Context, modID string) (*ficsit.AllVersionsResponse, error) {
	if p.Offline {
		return p.localProvider.ModVersionsWithDependencies(context, modID)
	}
	return p.ficsitProvider.ModVersionsWithDependencies(context, modID)
}

func (p MixedProvider) GetModName(context context.Context, modReference string) (*ficsit.GetModNameResponse, error) {
	if p.Offline {
		return p.localProvider.GetModName(context, modReference)
	}
	return p.ficsitProvider.GetModName(context, modReference)
}

func (p MixedProvider) IsOffline() bool {
	return p.Offline
}
