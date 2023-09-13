package cli

import (
	"context"
	"fmt"

	"github.com/Khan/genqlient/graphql"
	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type DependencyResolverError struct {
	pubgrub.SolvingError
	apiClient graphql.Client
}

func (e DependencyResolverError) Error() string {
	rootPkg := e.Cause().Terms()[0].Dependency()
	writer := pubgrub.NewStandardErrorWriter(rootPkg)
	writer.SetStringer(&DependencyResolverErrorStringer{
		apiClient:    e.apiClient,
		packageNames: make(map[string]string),
	})
	e.WriteTo(writer)
	return writer.String()
}

type DependencyResolverErrorStringer struct {
	apiClient    graphql.Client
	packageNames map[string]string
}

func (w *DependencyResolverErrorStringer) getPackageName(pkg string) string {
	if name, ok := w.packageNames[pkg]; ok {
		return name
	}
	result, err := ficsit.GetModName(context.Background(), w.apiClient, pkg)
	if err != nil {
		return pkg
	}
	w.packageNames[pkg] = result.Mod.Name
	return result.Mod.Name
}

func (w *DependencyResolverErrorStringer) Term(t pubgrub.Term, includeVersion bool) string {
	name := w.getPackageName(t.Dependency())
	fullName := fmt.Sprintf("%s (%s)", name, t.Dependency())
	if name == t.Dependency() {
		fullName = t.Dependency()
	}
	if includeVersion {
		if t.Constraint().IsAny() {
			return fmt.Sprintf("every version of %s", fullName)
		}
		return fmt.Sprintf("%s \"%s\"", fullName, t.Constraint())
	}
	return fullName
}
