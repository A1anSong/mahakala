package initialize

import (
	"github.com/go-resty/resty/v2"
	"server/global"
)

func Resty() {
	global.Resty = resty.New()
}
