package config

type Config struct {
	Envs []string `json:"envs"`

	Projects []string `json:"projects"`
}
