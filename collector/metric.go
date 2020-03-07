package collector

type Metric struct {
	Kind        string `json:"kind"`
	Group       string `json:"group"`
	Name        string `json:"name"`
	Units       string `json:"units"`
	Description string `json:"description"`
}

func NewMetric(name string, group string, description string) *Metric {
	return &Metric{
		Name:        name,
		Group:       group,
		Description: description,
	}
}
