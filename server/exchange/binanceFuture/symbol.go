package binanceFuture

type Symbol struct {
	Symbol       string   `json:"symbol"`
	ContractType string   `json:"contractType"`
	OnboardDate  int64    `json:"onboardDate"`
	Status       string   `json:"status"`
	Filters      []Filter `json:"filters"`
}
