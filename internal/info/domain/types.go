package domain

type Info struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	BuildTime   string `json:"build_time"`
}
