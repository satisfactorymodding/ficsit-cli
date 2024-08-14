package provider

import (
	"context"

	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type MixedProvider struct {
	onlineProvider  Provider
	offlineProvider Provider
	Offline         bool
}

func InitMixedProvider(onlineProvider Provider, offlineProvider Provider) *MixedProvider {
	return &MixedProvider{
		onlineProvider:  onlineProvider,
		offlineProvider: offlineProvider,
		Offline:         false,
	}
}

func (p MixedProvider) Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error) {
	if p.Offline {
		return p.offlineProvider.Mods(context, filter)
	}
	return p.onlineProvider.Mods(context, filter)
}

func (p MixedProvider) GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error) {
	if p.Offline {
		return p.offlineProvider.GetMod(context, modReference)
	}
	return p.onlineProvider.GetMod(context, modReference)
}

func (p MixedProvider) ModVersionsWithDependencies(context context.Context, modID string) ([]resolver.ModVersion, error) {
	if p.Offline {
		return p.offlineProvider.ModVersionsWithDependencies(context, modID) // nolint
	}
	return p.onlineProvider.ModVersionsWithDependencies(context, modID) // nolint
}

func (p MixedProvider) GetModName(context context.Context, modReference string) (*resolver.ModName, error) {
	if p.Offline {
		return p.offlineProvider.GetModName(context, modReference) // nolint
	}
	return p.onlineProvider.GetModName(context, modReference) // nolint
}

func (p MixedProvider) IsOffline() bool {
	return p.Offline
}
