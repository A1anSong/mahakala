package binanceFuture

import "time"

var Interval = map[string]time.Duration{
	"1m":  time.Minute * 1,
	"3m":  time.Minute * 3,
	"5m":  time.Minute * 5,
	"15m": time.Minute * 15,
	"30m": time.Minute * 30,
	"1h":  time.Hour * 1,
	"2h":  time.Hour * 2,
	"4h":  time.Hour * 4,
	"6h":  time.Hour * 6,
	"8h":  time.Hour * 8,
	"12h": time.Hour * 12,
	"1d":  time.Hour * 24,
	"3d":  time.Hour * 24 * 3,
	"1w":  time.Hour * 24 * 7,
	"1M":  time.Hour * 24 * 30,
}
