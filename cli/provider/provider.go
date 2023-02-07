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
	ResolveModDependencies(context context.Context, filter []ficsit.ModVersionConstraint) (*ficsit.ResolveModDependenciesResponse, error)
}
