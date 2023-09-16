package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/mircearoata/pubgrub-go/pubgrub"
	"github.com/mircearoata/pubgrub-go/pubgrub/semver"

	"github.com/satisfactorymodding/ficsit-cli/ficsit"
)

type DependencyResolverError struct {
	pubgrub.SolvingError
	apiClient   graphql.Client
	smlVersions []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion
	gameVersion int
}

func (e DependencyResolverError) Error() string {
	rootPkg := e.Cause().Terms()[0].Dependency()
	writer :=
		pubgrub.NewStandardErrorWriter(rootPkg).
			WithIncompatibilityStringer(
				MakeDependencyResolverErrorStringer(e.apiClient, e.smlVersions, e.gameVersion),
			)
	e.WriteTo(writer)
	return writer.String()
}

type DependencyResolverErrorStringer struct {
	pubgrub.StandardIncompatibilityStringer
	apiClient    graphql.Client
	smlVersions  []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion
	gameVersion  int
	packageNames map[string]string
}

func MakeDependencyResolverErrorStringer(apiClient graphql.Client, smlVersions []ficsit.SMLVersionsSmlVersionsGetSMLVersionsSml_versionsSMLVersion, gameVersion int) *DependencyResolverErrorStringer {
	s := &DependencyResolverErrorStringer{
		apiClient:    apiClient,
		smlVersions:  smlVersions,
		gameVersion:  gameVersion,
		packageNames: map[string]string{},
	}
	s.StandardIncompatibilityStringer = pubgrub.NewStandardIncompatibilityStringer().WithTermStringer(s)
	return s
}

func (w *DependencyResolverErrorStringer) getPackageName(pkg string) string {
	if pkg == "SML" {
		return "SML"
	}
	if pkg == "FactoryGame" {
		return "Satisfactory"
	}
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
		switch t.Dependency() {
		case "FactoryGame":
			// Remove ".0.0" from the versions mentioned, since only the major is ever used
			return fmt.Sprintf("%s \"%s\"", fullName, strings.ReplaceAll(t.Constraint().String(), ".0.0", ""))
		case "SML":
			var matched []semver.Version
			for _, v := range w.smlVersions {
				ver, err := semver.NewVersion(v.Version)
				if err != nil {
					// Assume it is contained in the constraint
					matched = append(matched, semver.Version{})
					continue
				}
				if t.Constraint().Contains(ver) {
					matched = append(matched, ver)
				}
			}
			if len(matched) == 1 {
				return fmt.Sprintf("%s \"%s\"", fullName, matched[0])
			}
			return fmt.Sprintf("%s \"%s\"", fullName, t.Constraint())
		default:
			res, err := ficsit.ModVersions(context.Background(), w.apiClient, t.Dependency(), ficsit.VersionFilter{
				Limit: 100,
			})
			if err != nil {
				return fmt.Sprintf("%s \"%s\"", fullName, t.Constraint())
			}
			var matched []semver.Version
			for _, v := range res.Mod.Versions {
				ver, err := semver.NewVersion(v.Version)
				if err != nil {
					// Assume it is contained in the constraint
					matched = append(matched, semver.Version{})
					continue
				}
				if t.Constraint().Contains(ver) {
					matched = append(matched, ver)
				}
			}
			if len(matched) == 1 {
				return fmt.Sprintf("%s \"%s\"", fullName, matched[0])
			}
			return fmt.Sprintf("%s \"%s\"", fullName, t.Constraint())
		}
	}
	return fullName
}

func (w *DependencyResolverErrorStringer) IncompatibilityString(incompatibility *pubgrub.Incompatibility, rootPkg string) string {
	terms := incompatibility.Terms()
	if len(terms) == 1 && terms[0].Dependency() == "FactoryGame" {
		return fmt.Sprintf("Satisfactory CL%d is installed", w.gameVersion)
	}
	return w.StandardIncompatibilityStringer.IncompatibilityString(incompatibility, rootPkg)
}
