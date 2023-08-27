package response

import "server/exchange"

type Exchange struct {
	exchange.Exchange
	Symbols []string `json:"symbols"`
}
