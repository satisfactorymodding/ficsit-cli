package cli

type ModVersion struct {
	ID           string
	Version      string
	Targets      map[string]VersionTarget
	Dependencies []VersionDependency
}

type VersionTarget struct {
	Link string
	Hash string
}

type VersionDependency struct {
	ModReference string
	Constraint   string
	Optional     bool
}
