package cli

type LockFile map[string]LockedMod

type LockedMod struct {
	Dependencies map[string]string `json:"dependencies"`
	Version      string            `json:"version"`
	Hash         string            `json:"hash"`
	Link         string            `json:"link"`
}

func (l LockFile) Clone() LockFile {
	lockFile := make(LockFile)
	for k, v := range l {
		lockFile[k] = v
	}
	return lockFile
}
