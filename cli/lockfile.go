package cli

type LockfileVersion int

const (
	InitialLockfileVersion = LockfileVersion(iota)

	ModTargetsLockfileVersion

	// Always last
	nextLockfileVersion
	CurrentLockfileVersion = nextLockfileVersion - 1
)

type LockFile struct {
	Mods    map[string]LockedMod `json:"mods"`
	Version LockfileVersion      `json:"version"`
}

type LockedMod struct {
	Dependencies map[string]string          `json:"dependencies"`
	Targets      map[string]LockedModTarget `json:"targets"`
	Version      string                     `json:"version"`
}

type LockedModTarget struct {
	Hash string `json:"hash"`
	Link string `json:"link"`
}

func MakeLockfile() *LockFile {
	return &LockFile{
		Mods:    make(map[string]LockedMod),
		Version: CurrentLockfileVersion,
	}
}

func (l *LockFile) Clone() *LockFile {
	lockFile := &LockFile{
		Mods:    make(map[string]LockedMod),
		Version: l.Version,
	}
	for k, v := range l.Mods {
		lockFile.Mods[k] = v
	}
	return lockFile
}

func (l *LockFile) Remove(modID ...string) *LockFile {
	for _, s := range modID {
		delete(l.Mods, s)
	}
	return l
}
