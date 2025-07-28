package config

type Config struct {
	OpenCmd string `json:"open_cmd"`

	Envs []string `json:"envs"`

	Projects []string `json:"projects"`
}
