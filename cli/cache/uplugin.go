package cache

type UPlugin struct {
	SemVersion   string    `json:"SemVersion"`
	FriendlyName string    `json:"FriendlyName"`
	Description  string    `json:"Description"`
	CreatedBy    string    `json:"CreatedBy"`
	GameVersion  string    `json:"GameVersion"`
	Plugins      []Plugins `json:"Plugins"`
}
type Plugins struct {
	Name       string `json:"Name"`
	SemVersion string `json:"SemVersion"`
	Enabled    bool   `json:"Enabled"`
	BasePlugin bool   `json:"BasePlugin"`
	Optional   bool   `json:"Optional"`
}
