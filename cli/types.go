package cli

type ModVersion struct {
	ID           string
	Version      string
	Dependencies []VersionDependency
}

type VersionDependency struct {
	ModReference string
	Constraint   string
	Optional     bool
}
