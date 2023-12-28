package provider

import (
	"context"

	resolver "github.com/satisfactorymodding/ficsit-resolver"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type Provider interface {
	resolver.Provider
	Mods(context context.Context, filter ficsit.ModFilter) (*ficsit.ModsResponse, error)
	GetMod(context context.Context, modReference string) (*ficsit.GetModResponse, error)
	IsOffline() bool
}
