package cli

type LockFile map[string]LockedMod

type LockedMod struct {
	Version      string            `json:"version"`
	Hash         string            `json:"hash"`
	Link         string            `json:"link"`
	Dependencies map[string]string `json:"dependencies"`
}

func (l LockFile) Clone() LockFile {
	lockFile := make(LockFile)
	for k, v := range l {
		lockFile[k] = v
	}
	return lockFile
}
