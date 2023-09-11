package model

type LeverageBracket struct {
	Symbol   string    `json:"symbol"`
	Brackets []Bracket `json:"brackets"`
}
