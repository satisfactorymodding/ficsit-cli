package cache

type UPlugin struct {
	SemVersion   string    `json:"SemVersion"`
	FriendlyName string    `json:"FriendlyName"`
	Description  string    `json:"Description"`
	CreatedBy    string    `json:"CreatedBy"`
	Plugins      []Plugins `json:"Plugins"`
}
type Plugins struct {
	Name       string `json:"Name"`
	Enabled    bool   `json:"Enabled"`
	BasePlugin bool   `json:"BasePlugin"`
	Optional   bool   `json:"Optional"`
	SemVersion string `json:"SemVersion"`
}
