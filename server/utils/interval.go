package utils

type Interval struct {
	Minutes int64
	String  string
}

var MapInterval = map[string]Interval{
	"1m":  {1, "1 minute"},
	"3m":  {3, "3 minutes"},
	"5m":  {5, "5 minutes"},
	"15m": {15, "15 minutes"},
	"30m": {30, "30 minutes"},
	"1h":  {60 * 1, "1 hour"},
	"2h":  {60 * 2, "2 hours"},
	"4h":  {60 * 4, "4 hours"},
	"6h":  {60 * 6, "6 hours"},
	"8h":  {60 * 8, "8 hours"},
	"12h": {60 * 12, "12 hours"},
	"1d":  {60 * 24, "1 day"},
	"3d":  {60 * 24 * 3, "3 days"},
	"5d":  {60 * 24 * 5, "5 days"},
	"1w":  {60 * 24 * 7, "1 week"},
	"1M":  {60 * 24 * 30, "1 month"},
}
