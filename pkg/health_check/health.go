package health_check

type Health struct {
	Checks []Checks `json:"checks"`
	Status string   `json:"status"`
}

type Checks struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
