package provider

import (
	"context"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type Provider interface {
	Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error)
	GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error)
	ModVersions(context context.Context, modReference string, filter ficsit.VersionFilter) (*ficsit.ModVersionsResponse, error)
	SMLVersions(context context.Context) (*ficsit.SMLVersionsResponse, error)
	ModVersionsWithDependencies(context context.Context, modID string) (*ficsit.AllVersionsResponse, error)
	GetModName(context context.Context, modReference string) (*ficsit.GetModNameResponse, error)
	IsOffline() bool
}
