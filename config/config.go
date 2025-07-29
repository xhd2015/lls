package config

type Config struct {
	OpenCmd string `json:"open_cmd"`

	HistoryFiles []string `json:"history_files"`

	Envs []string `json:"envs"`

	Projects []string `json:"projects"`
}
