package binanceFuture

import "time"

var Interval = map[string]time.Duration{
	"1m":  time.Minute * 1,
	"5m":  time.Minute * 5,
	"30m": time.Minute * 30,
	"1h":  time.Hour * 1,
	"2h":  time.Hour * 2,
	"4h":  time.Hour * 4,
	"6h":  time.Hour * 6,
	"12h": time.Hour * 12,
	"1d":  time.Hour * 24,
	"3d":  time.Hour * 24 * 3,
	"5d":  time.Hour * 24 * 5,
	"1w":  time.Hour * 24 * 7,
}
