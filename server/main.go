package main

import (
	"server/initialize"
)

func main() {
	initialize.Viper()
	initialize.Zap()
	initialize.Gorm()
	initialize.Exchange()
}
