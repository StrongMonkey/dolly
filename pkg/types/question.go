package types

type TemplateMeta struct {
	Name       string   `json:"name,omitempty"`
	Version    string   `json:"version,omitempty"`
	IconURL    string   `json:"iconUrl,omitempty"`
	Readme     string   `json:"readme,omitempty"`
	GoTemplate bool     `json:"goTemplate,omitempty"`
	Variables  []string `json:"variables,omitempty"`
	EnvSubst   bool     `json:"envSubst,omitempty"`
}
