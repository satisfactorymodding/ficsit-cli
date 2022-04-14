package cli

type LockFile map[string]LockedMod

type LockedMod struct {
	Version      string            `json:"version"`
	Dependencies map[string]string `json:"dependencies"`
}
